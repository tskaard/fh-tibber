package model

import (
	"github.com/futurehomeno/fimpgo/discovery"
	"github.com/futurehomeno/fimpgo/fimptype"
)

func GetDiscoveryResource() discovery.Resource {
	adInterfaces := []fimptype.Interface{
		{
			Type:      "in",
			MsgType:   "cmd.network.get_all_nodes",
			ValueType: "null",
			Version:   "1",
		}, {
			Type:      "in",
			MsgType:   "cmd.thing.get_inclusion_report",
			ValueType: "string",
			Version:   "1",
		}, {
		}, {
			Type:      "in",
			MsgType:   "cmd.thing.delete",
			ValueType: "string",
			Version:   "1",
		}, {
			Type:      "in",
			MsgType:   "cmd.system.get_connect_params",
			ValueType: "null",
			Version:   "1",
		}, {
			Type:      "in",
			MsgType:   "cmd.system.connect",
			ValueType: "str_map",
			Version:   "1",
		}, {
			Type:      "in",
			MsgType:   "cmd.system.disconnect",
			ValueType: "null",
			Version:   "1",
		}, {
			Type:      "in",
			MsgType:   "cmd.system.sync",
			ValueType: "null",
			Version:   "1",
		}, {
			Type:      "in",
			MsgType:   "cmd.stat.get_report",
			ValueType: "null",
			Version:   "1",
		}, {
			Type:      "out",
			MsgType:   "evt.thing.inclusion_report",
			ValueType: "object",
			Version:   "1",
		}, {
			Type:      "out",
			MsgType:   "evt.thing.exclusion_report",
			ValueType: "object",
			Version:   "1",
		}, {
			Type:      "out",
			MsgType:   "evt.network.all_nodes_report",
			ValueType: "object",
			Version:   "1",
		}, {
			Type:      "out",
			MsgType:   "evt.stat.report",
			ValueType: "string",
			Version:   "1",
		}, {
			Type:      "in",
			MsgType:   "evt.gateway.factory_reset",
			ValueType: "null",
			Version:   "1",
		}}

	adService := fimptype.Service{
		Name:             "tibber",
		Alias:            "Tibber managment",
		Address:          "/rt:ad/rn:tibber/ad:1",
		Enabled:          true,
		Interfaces:       adInterfaces,
	}
	return discovery.Resource{
		ResourceName:           "tibber",
		ResourceType:           discovery.ResourceTypeApp,
		ResourceFullName:       "Tibber",
		Description:            "Meter data from Tibber to futurehome",
		Author:                 "Tor Erik",
		IsInstanceConfigurable: false,
		InstanceId:             "1",
		Version:                "1",
		AdapterInfo: discovery.AdapterInfo{
			Technology:            "hue",
			FwVersion:             "all",
			NetworkManagementType: "inclusion_exclusion",
			Services:              []fimptype.Service{adService},
		},
	}

}
