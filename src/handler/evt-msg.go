package handler

import (
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
)

func (t *FimpTibberHandler) sendPowerMsg(addr string, power float64, oldMsg *fimpgo.FimpMessage) {
	service := "sensor_power"
	props := make(map[string]string)
	props["unit"] = "W"
	msg := fimpgo.NewMessage("evt.sensor.report", service, "float", power, props, nil, oldMsg)
	adr, _ := fimpgo.NewAddressFromString("pt:j1/mt:evt/rt:dev/rn:tibber/ad:1/sv:" + service + "/ad:" + addr)
	t.mqt.Publish(adr, msg)
	//log.Debug("Power message sent")
}

func (t *FimpTibberHandler) sendErrorReport(errString string, oldMsg *fimpgo.FimpMessage) {
	msg := fimpgo.NewStringMessage(
		"evt.error.report", "tibber",
		errString, nil, nil, oldMsg,
	)
	if err := t.mqt.RespondToRequest(oldMsg, msg); err == nil {
		log.WithError(err).Error("Could not publish MQTT message")
	}
}

func (t *FimpTibberHandler) sendConnectReport(status string, err string, oldMsg *fimpgo.FimpMessage) {
	connectReport := map[string]string{"status": status, "error": err}
	msg := fimpgo.NewStrMapMessage(
		"evt.system.connect_report", "tibber", connectReport, nil, nil, oldMsg,
	)
	if err := t.mqt.RespondToRequest(oldMsg, msg); err == nil {
		log.WithError(err).Error("Could not publish MQTT message")
	}
}

func (t *FimpTibberHandler) sendDisconnectReport(status string, err string, oldMsg *fimpgo.FimpMessage) {
	connectReport := map[string]string{"status": status, "error": err}
	msg := fimpgo.NewStrMapMessage(
		"evt.system.disconnect_report", "tibber", connectReport, nil, nil, oldMsg,
	)
	if err := t.mqt.RespondToRequest(oldMsg, msg); err == nil {
		log.WithError(err).Error("Could not publish MQTT message")
	}
}
