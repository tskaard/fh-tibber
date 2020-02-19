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
		Address:    "/rt:dev/rn:tibber/ad:1/sv:" + service + "/ad:" + addr,
		Name:       service,
		Groups:     []string{"ch_0"},
		Alias:      alias,
		Enabled:    true,
		Props:      props,
		Interfaces: sensorInterfaces,
	}
	return sensorService
}

func createMeterService(addr string, service string, alias string) fimptype.Service {
	props := make(map[string]interface{})
	props["sup_units"] = []string{"W"}
	props["sup_extended_vals"] = []string{
		"p_import", "e_import", "e_export",
		"last_e_import", "last_e_export",
		"p_import_min", "p_import_avg", "p_import_max",
		"p_export", "p_export_min", "p_export_max",
		"u1", "u2", "u3",
		"i1", "i2", "i3",
	}
	sensorService := fimptype.Service{
		Address: "/rt:dev/rn:tibber/ad:1/sv:" + service + "/ad:" + addr,
		Name:    service,
		Groups:  []string{"ch_0"},
		Alias:   alias,
		Enabled: true,
		Props:   props,
		Interfaces: []fimptype.Interface{
			createInterface("in", "cmd.meter.get_report", "null", "1"),
			createInterface("in", "cmd.meter.get_extended_report", "null", "1"),
			createInterface("out", "evt.meter.report", "float", "1"),
			createInterface("out", "evt.meter.extended_report", "float_map", "1"),
		},
	}
	return sensorService
}

func (t *FimpTibberHandler) sendInclusionReport(home tibber.Home, oldMsg *fimpgo.FimpMessage) {
	currentPrice, err := t.tibber.GetCurrentPrice(home.ID)
	if err != nil {
		log.Error("Cannot get prices from Tibber - ", err)
		return
	}

	priceSensorService := createSensorService(home.ID, "sensor_price", []string{currentPrice.Currency}, "price")
	meterService := createMeterService(home.ID, "meter_elec", "meter")

	incReort := fimptype.ThingInclusionReport{
		Address:        home.ID,
		CommTechnology: "tibber",
		ProductName:    "Tibber real time meter",
		Groups:         []string{"ch_0"},
		Services: []fimptype.Service{
			priceSensorService,
			meterService,
		},
		Alias:     home.AppNickname,
		ProductId: "HAN Solo",
		DeviceId:  home.ID,
	}

	msg := fimpgo.NewMessage("evt.thing.inclusion_report", "tibber", "object", incReort, nil, nil, oldMsg)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "tibber", ResourceAddress: "1"}
	t.mqt.Publish(&adr, msg)
	log.Debug("Inclusion report sent")
}
