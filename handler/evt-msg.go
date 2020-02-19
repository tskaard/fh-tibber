package handler

import (
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
)

func (t *FimpTibberHandler) sendSensorReportMsg(addr string, service string, value float64, unit string, oldMsg *fimpgo.FimpMessage) {
	props := make(map[string]string)
	props["unit"] = unit
	msg := fimpgo.NewMessage("evt.sensor.report", service, "float", value, props, nil, oldMsg)
	adr, _ := fimpgo.NewAddressFromString("pt:j1/mt:evt/rt:dev/rn:tibber/ad:1/sv:" + service + "/ad:" + addr)
	t.mqt.Publish(adr, msg)
}

func (t *FimpTibberHandler) sendMeterReportMsg(addr string, value float64, unit string, oldMsg *fimpgo.FimpMessage) {
	props := make(map[string]string)
	props["unit"] = unit
	msg := fimpgo.NewMessage("evt.meter.report", "meter_elec", "float", value, props, nil, oldMsg)
	adr, _ := fimpgo.NewAddressFromString("pt:j1/mt:evt/rt:dev/rn:tibber/ad:1/sv:meter/ad:" + addr)
	if err := t.mqt.Publish(adr, msg); err != nil {
		log.WithError(err).Error("Could not publish MQTT message")
	}
}

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
