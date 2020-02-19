package handler

import (
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
	tibber "github.com/tskaard/tibber-golang"
)

func (t *FimpTibberHandler) systemSync(oldMsg *fimpgo.Message) {
	if !t.state.Connected || t.state.AccessToken == "" {
		log.Error("Ad is not connected, not able to sync")
		return
	}
	t.sendInclusionReport(t.state.Home, oldMsg.Payload)
	log.Info("System synced")
}

func (t *FimpTibberHandler) systemDisconnect(msg *fimpgo.Message) {
	if !t.state.Connected {
		log.Error("Ad is not connected, no devices to exclude")
		return
	}

	t.sendExclusionReport(t.state.Home.ID, msg.Payload)
	if stream, ok := t.streams[t.state.Home.ID]; ok {
		stream.Stop()
		delete(t.streams, t.state.Home.ID)
	}
	_, err := t.tibber.SendPushNotification("Futurehome", t.state.Home.AppNickname+" is now disconnected from Futurehome")
	if err != nil {
		log.Debug("Push failed", err)
	}
	t.state.Connected = false
	t.state.AccessToken = ""
	t.state.Home = tibber.Home{}
	if err := t.db.Write("data", "state", t.state); err != nil {
		log.Error("Did not manage to write to file: ", err)
	}
}

func (t *FimpTibberHandler) systemGetConnectionParameter(oldMsg *fimpgo.Message) {
	// request api_key
	val := map[string]string{}
	if t.state.Connected {
		val["access_token"] = t.state.AccessToken
		val["home_id"] = t.state.Home.ID
	} else {
		val["access_token"] = "access_token"
		val["home_id"] = "home_id"
	}
	msg := fimpgo.NewStrMapMessage("evt.system.connect_params_report",
		"tibber", val, nil, nil, oldMsg.Payload)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter,
		ResourceName: "tibber", ResourceAddress: "1"}
	t.mqt.Publish(&adr, msg)
	log.Debug("Connect params message sent")
}

func (t *FimpTibberHandler) systemConnect(oldMsg *fimpgo.Message) {
	if t.state.Connected {
		log.Error("App is already connected with system")
		return
	}
	val, err := oldMsg.Payload.GetStrMapValue()
	if err != nil {
		log.Error("Wrong payload type , expected StrMap")
		return
	}
	if val["access_token"] == "" {
		log.Error("Did not get a security_key")
		return
	}

	t.tibber.Token = val["access_token"]

	// If home id is specified, connect to it. Otherwise connect to first home with RealTimeConsumptionEnabled
	var homeId = val["home_id"]
	if homeId != "" && homeId != "home_id" {
		home, err := t.tibber.GetHomeById(homeId)
		if err != nil {
			log.Error("Cannot get home by id from Tibber - ", err)
			t.tibber.Token = ""
			return
		}

		if home.Features.RealTimeConsumptionEnabled {
			t.startSubscriptionForHome(oldMsg, &home)
			return
		}
	} else {
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
				t.startSubscriptionForHome(oldMsg, &home)
				return
			}
		}
	}

	log.Warning("Could not find home with real time consumption device")
	t.state.AccessToken = ""
	t.state.Connected = false

	if err := t.db.Write("data", "state", t.state); err != nil {
		log.Error("Did not manage to write to file: ", err)
		return
	}

	t.sendConnectReport("error",
		"Could not find home with real time consumption device", oldMsg.Payload)
}

func (t *FimpTibberHandler) startSubscriptionForHome(oldMsg *fimpgo.Message, home *tibber.Home) {
	t.sendInclusionReport(*home, oldMsg.Payload)
	stream := tibber.NewStream(home.ID, t.tibber.Token)
	stream.StartSubscription(t.tibberMsgCh)
	t.streams[home.ID] = stream
	_, err := t.tibber.SendPushNotification("Futurehome", home.AppNickname+" is now connected to Futurehome 🎉")
	if err != nil {
		log.Debug("Push failed", err)
	}

	t.state.AccessToken = t.tibber.Token
	t.state.Connected = true
	t.state.Home = *home

	if err := t.db.Write("data", "state", t.state); err != nil {
		log.Error("Did not manage to write to file: ", err)
		return
	}
	log.Debug("System connected")
	t.sendConnectReport("ok", "", oldMsg.Payload)
}

func (t *FimpTibberHandler) thingInclusionReport(msg *fimpgo.Message) {
	if !t.state.Connected {
		log.Error("Ad is not connected, not able to sync")
		return
	}
	id, err := msg.Payload.GetStringValue()
	if err != nil {
		log.Error("Wrong payload type , expected String")
		return
	}
	if t.state.Home.ID == id {
		t.sendInclusionReport(t.state.Home, msg.Payload)
		log.WithField("id", id).Info("Inclusion report sent")
	} else {
		t.sendErrorReport("NOT_FOUND", msg.Payload)
	}
}

func (t *FimpTibberHandler) thingDelete(msg *fimpgo.Message) {
	if !t.state.Connected {
		log.Error("Ad is not connected, not able to sync")
		return
	}
	id, err := msg.Payload.GetStringValue()
	if err != nil {
		log.Error("Wrong payload type , expected String")
		return
	}
	if t.state.Home.ID == id {
		t.sendExclusionReport(t.state.Home.ID, msg.Payload)
		if stream, ok := t.streams[t.state.Home.ID]; ok {
			stream.Stop()
			delete(t.streams, t.state.Home.ID)
		}
		t.state.Home = tibber.Home{}
		log.WithField("id", id).Info("Inclusion report sent")
	} else {
		t.sendErrorReport("NOT_FOUND", msg.Payload)
	}

}
