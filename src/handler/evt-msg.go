package handler

import (
	"github.com/futurehomeno/fimpgo"
)

func (t *FimpTibberHandler) sendPowerMsg(addr string, power float64, oldMsg *fimpgo.FimpMessage) {
	service := "sensor_power"
	props := make(map[string]string)
	props["unit"] = "W"
	msg := fimpgo.NewMessage("evt.sensor.report", service, "float", power, props, nil, oldMsg)
	adr, _ := fimpgo.NewAddressFromString("pt:j1/mt:evt/rt:dev/rn:fh-tibber/ad:1/sv:" + service + "/ad:" + addr)
	t.mqt.Publish(adr, msg)
	//log.Debug("Power message sent")
}
