package model

import (
	tibber "github.com/tskaard/tibber-golang"
)

// State is the internal state of the app
type State struct {
	Connected   bool        `json:"connected"`
	AccessToken string      `json:"accessToken"`
	Home        tibber.Home `json:"home"`
}
