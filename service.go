package main

import (
	"flag"
	"fmt"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/discovery"
	"github.com/futurehomeno/fimpgo/edgeapp"
	log "github.com/sirupsen/logrus"
	"github.com/tskaard/fh-tibber/handler"
	"github.com/tskaard/fh-tibber/model"
	"github.com/tskaard/fh-tibber/utils"
	"time"
)

func main() {
	var workDir string
	flag.StringVar(&workDir, "c", "", "Work dir")
	flag.Parse()
	if workDir == "" {
		workDir = "./"
	} else {
		fmt.Println("Work dir ", workDir)
	}
	appLifecycle := edgeapp.NewAppLifecycle()

	configs := model.NewConfigs(workDir)
	err := configs.LoadFromFile()
	if err != nil {
		appLifecycle.SetAppState(edgeapp.AppStateStartupError, nil)
		fmt.Print(err)
		panic("Can't load config file.")

	}
	if err != nil {
		appLifecycle.SetConfigState(edgeapp.ConfigStateNotConfigured)
		log.Debug("Not able to load state")
		log.Debug(err)
	}

	utils.SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting Tibber----------------")
	appLifecycle.SetAppState(edgeapp.AppStateStarting, nil)
	appLifecycle.SetAuthState(edgeapp.AuthStateNotAuthenticated)
	appLifecycle.SetConfigState(edgeapp.ConfigStateNotConfigured)
	appLifecycle.SetConnectionState(edgeapp.ConnStateDisconnected)

	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI, configs.MqttClientIdPrefix, configs.MqttUsername, configs.MqttPassword, true, 1, 1)
	err = mqtt.Start()
	responder := discovery.NewServiceDiscoveryResponder(mqtt)
	responder.RegisterResource(model.GetDiscoveryResource())
	responder.Start()
	if err != nil {
		log.Error("Can't connect to broker. Error:", err.Error())
	} else {
		log.Info("--------------Connected----------------")
	}
	defer mqtt.Stop()

	if err := edgeapp.NewSystemCheck().WaitForInternet(5 * time.Minute); err == nil {
		log.Info("<main> Internet connection - OK")
	} else {
		log.Error("<main> Internet connection - ERROR")
	}

	tibberHandler := handler.NewTibberHandler(mqtt, appLifecycle)

	if configs.AccessToken == "" && configs.HomeID == "" {
		log.Info("<main> Token is not set. The app is not configured")
		appLifecycle.SetAppState(edgeapp.AppStateNotConfigured, nil)
	} else {
		appLifecycle.SetConfigState(edgeapp.ConfigStateConfigured)
		appLifecycle.SetAuthState(edgeapp.AuthStateAuthenticated)
		err := tibberHandler.Start(configs.AccessToken, configs.HomeID)
		appLifecycle.SetAppState(edgeapp.AppStateRunning, nil)
		if err != nil {
			// Handle error
			appLifecycle.SetConnectionState(edgeapp.ConnStateDisconnected)
			appLifecycle.SetLastError(fmt.Sprint("Can't connect to Tibber api . Err:",err.Error()))
		}else {
			appLifecycle.SetConnectionState(edgeapp.ConnStateConnected)
		}
	}

	fimpHandler := handler.NewFimpTibberHandler(mqtt, appLifecycle, tibberHandler, configs)
	fimpHandler.Start()

	for {
		appLifecycle.WaitForState("main", edgeapp.AppStateRunning)
		select {}
	}

}
