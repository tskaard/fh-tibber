package handler

import (
	"path/filepath"

	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/edgeapp"
	"github.com/futurehomeno/fimpgo/utils"
	log "github.com/sirupsen/logrus"
	"github.com/tskaard/fh-tibber/model"
	tibber "github.com/tskaard/tibber-golang"
)

// FimpTibberHandler structure
type FimpTibberHandler struct {
	inboundMsgCh fimpgo.MessageCh
	mqt          *fimpgo.MqttTransport
	tibber       *TibberHandler
	appLifecycle *edgeapp.Lifecycle
	configs      *model.Configs
	env          string
}

// NewFimpTibberHandler construct new handler
func NewFimpTibberHandler(transport *fimpgo.MqttTransport, appLifecycle *edgeapp.Lifecycle, tibber *TibberHandler, configs *model.Configs) *FimpTibberHandler {
	t := &FimpTibberHandler{
		inboundMsgCh: make(fimpgo.MessageCh, 5),
		mqt:          transport,
		appLifecycle: appLifecycle,
		tibber:       tibber,
		configs:      configs,
	}
	t.mqt.RegisterChannel("ch1", t.inboundMsgCh)
	hubInfo, err := utils.NewHubUtils().GetHubInfo()
	if err == nil && hubInfo != nil {
		t.env = hubInfo.Environment
	} else {
		t.env = utils.EnvProd
	}
	return t
}

// Start handler
func (t *FimpTibberHandler) Start() error {
	t.mqt.Subscribe("pt:j1/mt:cmd/rt:dev/rn:tibber/ad:1/#")
	t.mqt.Subscribe("pt:j1/mt:cmd/rt:ad/rn:tibber/ad:1")
	// Listen for the factory reset event
	t.mqt.Subscribe("pt:j1/mt:evt/rt:ad/rn:gateway/ad:1")

	var errr error
	go func(msgChan fimpgo.MessageCh) {
		for {
			select {
			case newMsg := <-msgChan:
				t.routeFimpMessage(newMsg)
			}
		}
	}(t.inboundMsgCh)
	return errr
}

func (t *FimpTibberHandler) routeFimpMessage(newMsg *fimpgo.Message) {
	log.WithField("type", newMsg.Payload.Type).Debug("New fimp msg")
	switch newMsg.Payload.Service {
	case "sensor_price":
		switch newMsg.Payload.Type {
		case "cmd.sensor.get_report":
			currentPrice, err := t.tibber.client.GetCurrentPrice(t.tibber.home.ID)
			if err != nil {
				log.Error("Cannot get prices from Tibber - ", err)
				return
			}
			t.tibber.sendSensorReportMsg(t.tibber.home.ID, "sensor_price", currentPrice.Total, currentPrice.Currency, newMsg.Payload)
			log.Debug("sensor_price sent")
		}
	case "meter_elec":
		switch newMsg.Payload.Type {
		case "cmd.meter.get_report":
			log.Debug("cmd.meter.get_report requested but not implemented")
		case "cmd.meter_ext.get_report":
			log.Debug("cmd.meter_ext.get_report requested but not implemented")
		case "cmd.meter.reset":
			log.Debug("cmd.meter.reset requested but not implemented")
		}

	case "tibber":
		adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "tibber", ResourceAddress: "1"}
		switch newMsg.Payload.Type {

		case "cmd.system.disconnect":
			t.systemDisconnect(newMsg)

		case "cmd.auth.set_tokens":
			log.Info("Configuring tokens")
			newTokens := AuthData{}
			err := newMsg.Payload.GetObjectValue(&newTokens)
			if err != nil {
				log.Error("Incorrect login message ")
				return
			}
			if newTokens.AccessToken != "" {
				t.configs.AccessToken = newTokens.AccessToken
				t.tibber.client.Token = newTokens.AccessToken
				t.tibber.stream.Token = newTokens.AccessToken

				// Getting homes
				homes, err := t.tibber.client.GetHomes()
				if err != nil {
					log.Error("Cannot get homes from Tibber - ", err)
					t.tibber.client.Token = ""
					break
				}
				for _, home := range homes {
					log.Debug(home.ID)
					if home.ID == "" {
						break
					}
					if home.Features.RealTimeConsumptionEnabled {
						t.configs.HomeID = home.ID
						t.tibber.stream.ID = home.ID
						t.tibber.home = &home
						break
					}
				}
				var status string
				errStr := ""
				if t.tibber.stream.ID != "" {
					//t.tibber.stream.StartSubscription(t.tibber.msgChan)
					t.tibber.Start(t.tibber.client.Token, t.tibber.stream.ID)

					t.tibber.client.SendPushNotification("Futurehome", t.tibber.home.AppNickname+" is now connected to Futurehome 🎉")
					t.configs.SaveToFile()

					t.appLifecycle.SetAppState(edgeapp.AppStateRunning, nil)
					t.appLifecycle.SetConfigState(edgeapp.ConfigStateConfigured)
					t.appLifecycle.SetConnectionState(edgeapp.ConnStateConnected)
					t.appLifecycle.SetAuthState(edgeapp.AuthStateAuthenticated)
					status = edgeapp.AuthStateAuthenticated
					// Send inc report
					t.sendInclusionReport(*t.tibber.home, newMsg.Payload)
				} else {
					log.Info("Tokens configuration failed with error : ", err)
					t.appLifecycle.SetAppState(edgeapp.AppStateNotConfigured, nil)
					t.appLifecycle.SetConfigState(edgeapp.ConfigStateNotConfigured)
					t.appLifecycle.SetConnectionState(edgeapp.ConnStateConnected)
					t.appLifecycle.SetAuthState(edgeapp.AuthStateNotAuthenticated)
					status = edgeapp.AuthStateNotAuthenticated
					if err != nil {
						t.appLifecycle.SetLastError(err.Error())
						errStr = err.Error()
					}
				}
				val := edgeapp.AuthResponse{
					Status:    status,
					ErrorText: errStr,
					ErrorCode: "",
				}
				msg := fimpgo.NewMessage("evt.auth.status_report", "tibber", fimpgo.VTypeObject, val, nil, nil, newMsg.Payload)
				if err := t.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
					// if response topic is not set , sending back to default application event topic
					t.mqt.Publish(adr, msg)
				}
			} else {
				log.Error("Empty tokens , message was skipped")
			}

		case "cmd.app.get_manifest":
			log.Info("Manifest request")
			mode, err := newMsg.Payload.GetStringValue()
			if err != nil {
				log.Error("Incorrect request format ")
				return
			}
			manifest := edgeapp.NewManifest()
			err = manifest.LoadFromFile(filepath.Join(t.configs.GetDefaultDir(), "app-manifest.json"))
			if err != nil {
				log.Error("Failed to load manifest file .Error :", err.Error())
				return
			}
			if mode == "manifest_state" {
				manifest.AppState = *t.appLifecycle.GetAllStates()
				manifest.AppState.Auth = string(t.appLifecycle.AuthState())
				confState := model.PublicConfigs{}
				confState.ConnectionState = string(t.appLifecycle.ConnectionState())
				confState.Errors = t.appLifecycle.GetAllStates().LastErrorText
				manifest.ConfigState = confState
			}
			if t.env == utils.EnvBeta {
				manifest.Auth.AuthEndpoint = "https://partners-beta.futurehome.io/api/control/edge/proxy/auth-code"
				manifest.Auth.RedirectURL = "https://app-static-beta.futurehome.io/playground_oauth_callback"
				manifest.Auth.CodeGrantLoginPageUrl = "https://thewall.tibber.com/connect/authorize?client_id=8nr3zyLa-dF-qIcCtXET0sq9xCxK6EjCKn7jx3A9GY8&redirect_uri=https://app-static-beta.futurehome.io/playground_oauth_callback&response_type=code&scope=tibber_graph"
			} else {
				manifest.Auth.AuthEndpoint = "https://partners.futurehome.io/api/control/edge/proxy/auth-code"
				manifest.Auth.RedirectURL = "https://app-static.futurehome.io/playground_oauth_callback"
				manifest.Auth.CodeGrantLoginPageUrl = "https://thewall.tibber.com/connect/authorize?client_id=8nr3zyLa-dF-qIcCtXET0sq9xCxK6EjCKn7jx3A9GY8&redirect_uri=https://app-static.futurehome.io/playground_oauth_callback&response_type=code&scope=tibber_graph"
			}

			msg := fimpgo.NewMessage("evt.app.manifest_report", "tibber", fimpgo.VTypeObject, manifest, nil, nil, newMsg.Payload)
			if err := t.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				// if response topic is not set , sending back to default application event topic
				t.mqt.Publish(adr, msg)
			}

		case "cmd.app.get_state":
			msg := fimpgo.NewMessage("evt.app.manifest_report", "tibber", fimpgo.VTypeObject, t.appLifecycle.GetAllStates(), nil, nil, newMsg.Payload)
			if err := t.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				// if response topic is not set , sending back to default application event topic
				t.mqt.Publish(adr, msg)
			}

		case "cmd.auth.logout":
			// exclude device
			t.tibber.stream.Stop()
			t.sendExclusionReport(t.tibber.home.ID, newMsg.Payload)

			t.tibber.home = &tibber.Home{}
			t.configs.HomeID = ""
			t.configs.SaveToFile()

			if t.configs.HomeID == "" {
				// set appLifeCycle values
				t.appLifecycle.SetAppState(edgeapp.AppStateNotConfigured, nil)
				t.appLifecycle.SetConfigState(edgeapp.ConfigStateNotConfigured)
				t.appLifecycle.SetConnectionState(edgeapp.ConnStateDisconnected)
				t.appLifecycle.SetAuthState(edgeapp.AuthStateNotAuthenticated)

				// respond to topic with necessary value(s)
				val := map[string]interface{}{
					"errors":  nil,
					"success": true,
				}
				msg := fimpgo.NewMessage("evt.pd7.response", "vinculum", fimpgo.VTypeObject, val, nil, nil, newMsg.Payload)
				if err := t.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
					log.Error("Could not respont to wanted request")
				}
				log.Info("Logged out successfully")
			} else {
				log.Error("Something went wrong when logging out")
			}

		case "cmd.config.get_extended_report":
			msg := fimpgo.NewMessage("evt.config.extended_report", "tibber", fimpgo.VTypeObject, t.configs, nil, nil, newMsg.Payload)
			if err := t.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				t.mqt.Publish(adr, msg)
			}

		case "cmd.system.sync":
			t.systemSync(newMsg)

		case "cmd.network.get_all_nodes":
		// TODO: Send information about all devices

		case "cmd.thing.get_inclusion_report":
			t.thingInclusionReport(newMsg)

		case "cmd.thing.delete":
			t.thingDelete(newMsg)

			// case "evt.gateway.factory_reset":
			// 	t.systemDisconnect(newMsg)

		}

	}
}
