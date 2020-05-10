package model

import (
	"github.com/futurehomeno/fimpgo/discovery"
)

// GetDiscoveryResource contains the discovery response
func GetDiscoveryResource() discovery.Resource {

	return discovery.Resource{
		ResourceName:           "tibber",
		ResourceType:           discovery.ResourceTypeAd,
		ResourceFullName:       "Tibber",
		Description:            "Meter data from Tibber to Futurehome",
		Author:                 "tor.erik@futurehome.no",
		IsInstanceConfigurable: false,
		InstanceId:             "1",
		Version:                "1",
		AdapterInfo: discovery.AdapterInfo{
			Technology:            "tibber",
			FwVersion:             "all",
			NetworkManagementType: "inclusion_exclusion",
		},
	}

}
