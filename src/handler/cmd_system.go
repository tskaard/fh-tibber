package handler

import (
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
	tibber "github.com/tskaard/tibber-golang"
)

func (t *FimpTibberHandler) systemSync(oldMsg *fimpgo.Message) {
	log.Debug("cmd.system.sync")
	if !t.state.Connected || t.state.AccessToken == "" {
		log.Error("Ad is not connected, not able to sync")
		return
	}
	for _, home := range t.state.Homes {
		t.sendInclusionReport(home, oldMsg.Payload)
	}
	log.Info("System synced")
}

func (t *FimpTibberHandler) systemDisconnect(msg *fimpgo.Message) {
	log.Debug("cmd.system.disconnect")
	if !t.state.Connected {
		log.Error("Ad is not connected, no devices to exclude")
		return
	}
	for _, home := range t.state.Homes {
		t.sendExclusionReport(home.ID, msg.Payload)
		if stream, ok := t.streams[home.ID]; ok {
			stream.Stop()
			delete(t.streams, home.ID)
		}
		_, err := t.tibber.SendPushNotification("Futurehome", home.AppNickname+" is now disconnected from Futurehome")
		if err != nil {
			log.Debug("Push failed", err)
		}
	}
	t.state.Connected = false
	t.state.AccessToken = ""
	t.state.Homes = nil
	if err := t.db.Write("data", "state", t.state); err != nil {
		log.Error("Did not manage to write to file: ", err)
	}
}

func (t *FimpTibberHandler) systemGetConnectionParameter(oldMsg *fimpgo.Message) {
	log.Debug("cmd.system.get_connect_params")
	// request api_key
	val := map[string]string{
		"address": "api.tibber.com",
		"id":      "tibber",
	}
	if t.state.Connected {
		val["security_key"] = t.state.AccessToken
	} else {
		val["security_key"] = "api_key"
	}
	msg := fimpgo.NewStrMapMessage("evt.system.connect_params_report", "tibber", val, nil, nil, oldMsg.Payload)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "tibber", ResourceAddress: "1"}
	t.mqt.Publish(&adr, msg)
	log.Debug("Connect params message sent")
}

func (t *FimpTibberHandler) systemConnect(oldMsg *fimpgo.Message) {
	log.Debug("cmd.system.connect")
	if t.state.Connected {
		log.Error("App is already connected with system")
		return
	}
	val, err := oldMsg.Payload.GetStrMapValue()
	if err != nil {
		log.Error("Wrong payload type , expected StrMap")
		return
	}
	if val["security_key"] == "" {
		log.Error("Did not get a security_key")
		return
	}

	t.tibber.Token = val["security_key"]
	homes, err := t.tibber.GetHomes()
	if err != nil {
		log.Error("Cannot get homes from Tibber - ", err)
		t.tibber.Token = ""
		return
	}

	for _, home := range homes {
		log.Debug(home.ID)
		if home.ID == "" {
			return
		}
		if home.Features.RealTimeConsumptionEnabled {
			t.state.Homes = append(t.state.Homes, home)
			t.sendInclusionReport(home, oldMsg.Payload)
			stream := tibber.NewStream(home.ID, t.tibber.Token)
			stream.StartSubscription(t.tibberMsgCh)
			t.streams[home.ID] = stream
			_, err := t.tibber.SendPushNotification("Futurehome", home.AppNickname+" is now connected to Futurehome 🎉")
			if err != nil {
				log.Debug("Push failed", err)
			}
			// Connect to pulse and start subscription
		}
	}

	if t.state.Homes != nil {
		t.state.AccessToken = val["security_key"]
		t.state.Connected = true
	} else {
		t.state.AccessToken = ""
		t.state.Connected = false
	}

	if err := t.db.Write("data", "state", t.state); err != nil {
		log.Error("Did not manage to write to file: ", err)
		return
	}
	log.Debug("System connected")
}
