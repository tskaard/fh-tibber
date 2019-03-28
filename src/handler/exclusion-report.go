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
	msg := fimpgo.NewMessage("evt.thing.exclusion_report", "fh-tibber", "object", exReport, nil, nil, oldMsg)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "fh-tibber", ResourceAddress: "1"}
	t.mqt.Publish(&adr, msg)
	log.Debug("Exclusion report sent")
}
