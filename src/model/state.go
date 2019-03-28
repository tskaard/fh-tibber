package model

import (
	tibber "github.com/tskaard/tibber-golang"
)

type State struct {
	Connected   bool          `json:"connected"`
	AccessToken string        `json:"accessToken"`
	Homes       []tibber.Home `json:"homes"`
}
