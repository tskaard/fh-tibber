package handler

import (
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/fimptype"
	log "github.com/sirupsen/logrus"
)

func (t *FimpTibberHandler) sendExclusionReport(addr string, oldMsg *fimpgo.FimpMessage) {
	exReport := fimptype.ThingExclusionReport{
		Address: addr,
	}
	msg := fimpgo.NewMessage(
		"evt.thing.exclusion_report", "tibber",
		"object", exReport, nil, nil, oldMsg,
	)
	if err := t.mqt.RespondToRequest(oldMsg, msg); err == nil {
		log.WithError(err).Error("Could not publish MQTT message")
	}
	log.Debug("Exclusion report sent")
}
