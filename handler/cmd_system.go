package handler

import (
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/edgeapp"
	log "github.com/sirupsen/logrus"
	tibber "github.com/tskaard/tibber-golang"
)

func (t *FimpTibberHandler) systemSync(oldMsg *fimpgo.Message) {
	if t.appLifecycle.ConfigState() != edgeapp.ConfigStateConfigured {
		log.Error("Tibber is not configured, not able to sync")
		return
	}
	t.sendInclusionReport(*t.tibber.home, oldMsg.Payload)
	val := edgeapp.ButtonActionResponse{
		Operation:       "cmd.system.sync",
		OperationStatus: "ok",
		Next:            "reload",
		ErrorCode:       "",
		ErrorText:       "",
	}
	adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "tibber", ResourceAddress: "1"}
	msg := fimpgo.NewMessage("evt.app.config_action_report", "tibber", fimpgo.VTypeObject, val, nil, nil, oldMsg.Payload)
	if err := t.mqt.RespondToRequest(oldMsg.Payload, msg); err != nil {
		t.mqt.Publish(adr, msg)
	}
	log.Info("System synced")

}

func (t *FimpTibberHandler) systemDisconnect(oldMsg *fimpgo.Message) {
	if t.appLifecycle.ConfigState() != edgeapp.ConfigStateConfigured {
		log.Error("Tibber is not configured, not able to sync, no devices to exclude")
		t.sendDisconnectReport("error", "Adapter is not connected", oldMsg.Payload)
		return
	}
	t.tibber.stream.Stop()

	t.sendExclusionReport(t.tibber.home.ID, oldMsg.Payload)

	_, err := t.tibber.client.SendPushNotification("Futurehome", t.tibber.home.AppNickname+" is now disconnected from Futurehome")
	if err != nil {
		log.Debug("Push failed", err)
	}
	// TODO: delete config and state. Do this by setting not configured or something and doing it in main?
	// Change app state, connection state, auth state...

	t.configs.AccessToken = ""
	t.configs.HomeID = ""
	t.configs.SaveToFile()
	t.tibber.home = &tibber.Home{}
	t.tibber.client.Token = ""
	t.tibber.stream.ID = ""
	t.tibber.stream.Token = ""
	t.appLifecycle.SetAppState(edgeapp.AppStateNotConfigured, nil)
	t.appLifecycle.SetAuthState(edgeapp.AuthStateNotAuthenticated)
	t.appLifecycle.SetConfigState(edgeapp.ConfigStateNotConfigured)
	t.appLifecycle.SetConnectionState(edgeapp.ConnStateConnected)

	t.sendDisconnectReport("ok", "", oldMsg.Payload)
}

func (t *FimpTibberHandler) thingInclusionReport(msg *fimpgo.Message) {
	if t.appLifecycle.ConfigState() != edgeapp.ConfigStateConfigured {
		log.Error("Ad is not connected, not able to sync")
		return
	}
	id, err := msg.Payload.GetStringValue()
	if err != nil {
		log.Error("Wrong payload type , expected String")
		return
	}
	if t.tibber.home.ID == id {
		t.sendInclusionReport(*t.tibber.home, msg.Payload)
		log.WithField("id", id).Info("Inclusion report sent")
	} else {
		t.sendErrorReport("NOT_FOUND", msg.Payload)
	}
}

func (t *FimpTibberHandler) thingDelete(msg *fimpgo.Message) {
	if t.appLifecycle.ConfigState() != edgeapp.ConfigStateConfigured {
		log.Error("Ad is not connected, not able to sync")
		return
	}
	id, err := msg.Payload.GetStringValue()
	if err != nil {
		log.Error("Wrong payload type , expected String")
		return
	}
	if t.tibber.home.ID == id {
		t.tibber.stream.Stop()
		t.sendExclusionReport(t.tibber.home.ID, msg.Payload)

		t.tibber.home = &tibber.Home{}
		t.configs.HomeID = ""
		t.configs.SaveToFile()
		log.WithField("id", id).Info("Inclusion report sent")
	} else {
		t.sendErrorReport("NOT_FOUND", msg.Payload)
	}
}
