package handler

import (
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
)

func (t *FimpTibberHandler) sendErrorReport(errString string, oldMsg *fimpgo.FimpMessage) {
	msg := fimpgo.NewStringMessage(
		"evt.error.report", "tibber",
		errString, nil, nil, oldMsg,
	)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter,
		ResourceName: "tibber", ResourceAddress: "1"}
	if err := t.mqt.Publish(&adr, msg); err != nil {
		log.WithError(err).Error("Could not publish MQTT message")
	}
}

func (t *FimpTibberHandler) sendConnectReport(status string, err string, oldMsg *fimpgo.FimpMessage) {
	connectReport := map[string]string{"status": status, "error": err}
	msg := fimpgo.NewStrMapMessage(
		"evt.system.connect_report", "tibber", connectReport, nil, nil, oldMsg,
	)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter,
		ResourceName: "tibber", ResourceAddress: "1"}
	if err := t.mqt.Publish(&adr, msg); err != nil {
		log.WithError(err).Error("Could not publish MQTT message")
	}
}

func (t *FimpTibberHandler) sendDisconnectReport(status string, err string, oldMsg *fimpgo.FimpMessage) {
	connectReport := map[string]string{"status": status, "error": err}
	msg := fimpgo.NewStrMapMessage(
		"evt.system.disconnect_report", "tibber", connectReport, nil, nil, oldMsg,
	)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter,
		ResourceName: "tibber", ResourceAddress: "1"}
	if err := t.mqt.Publish(&adr, msg); err != nil {
		log.WithError(err).Error("Could not publish MQTT message")
	}
}
