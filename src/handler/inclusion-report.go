package handler

import (
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/fimptype"
	log "github.com/sirupsen/logrus"
	tibber "github.com/tskaard/tibber-golang"
)

func createInterface(iType string, msgType string, valueType string, version string) fimptype.Interface {
	inter := fimptype.Interface{
		Type:      iType,
		MsgType:   msgType,
		ValueType: valueType,
		Version:   version,
	}
	return inter
}

func createSensorService(addr string, service string, supUnits []string, alias string) fimptype.Service {
	cmdSensorGetReport := createInterface("in", "cmd.sensor.get_report", "null", "1")
	evtSensorReport := createInterface("out", "evt.sensor.report", "float", "1")
	sensorInterfaces := []fimptype.Interface{}
	sensorInterfaces = append(sensorInterfaces, cmdSensorGetReport, evtSensorReport)

	props := make(map[string]interface{})
	props["sup_units"] = supUnits
	sensorService := fimptype.Service{
		Address:    "/rt:dev/rn:fh-tibber/ad:1/sv:" + service + "/ad:" + addr,
		Name:       service,
		Groups:     []string{"ch_0"},
		Alias:      alias,
		Enabled:    true,
		Props:      props,
		Interfaces: sensorInterfaces,
	}
	return sensorService
}

func (t *FimpTibberHandler) sendInclusionReport(home tibber.Home, oldMsg *fimpgo.FimpMessage) {

	powerSensorService := createSensorService(home.ID, "sensor_power", []string{"W"}, "power")

	services := []fimptype.Service{}
	services = append(services, powerSensorService)
	incReort := fimptype.ThingInclusionReport{
		Address:        home.ID,
		CommTechnology: "wss",
		ProductName:    "Tibber Pulse",
		Groups:         []string{"ch_0"},
		Services:       services,
		Alias:          home.AppNickname,
		ProductId:      "HAN Solo",
		DeviceId:       home.MeteringPointData.ConsumptionEan,
	}

	msg := fimpgo.NewMessage("evt.thing.inclusion_report", "fh-tibber", "object", incReort, nil, nil, oldMsg)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "fh-tibber", ResourceAddress: "1"}
	t.mqt.Publish(&adr, msg)
	log.Debug("Inclusion report sent")
}
