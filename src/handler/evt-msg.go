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
	adr := fimpgo.Address{
		MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter,
		ResourceName: "tibber", ResourceAddress: "1",
	}
	if err := t.mqt.Publish(&adr, msg); err != nil {
		log.WithError(err).Error("Could not publish MQTT message")
	}
	log.Debug("Inclusion report sent")
}
