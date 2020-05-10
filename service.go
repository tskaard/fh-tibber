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

	configs := edgeapp.NewConfigs(workDir)
	err := configs.LoadFromFile()
	if err != nil {
		fmt.Print(err)
		panic("Can't load config file.")
	}

	utils.SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting Tibber----------------")
	appLifecycle.PublishSystemEvent(edgeapp.EventConfiguring, "main", nil)

	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI, configs.MqttClientIdPrefix, configs.MqttUsername, configs.MqttPassword, true, 1, 1)
	err = mqtt.Start()
	if err != nil {
		log.Error("Can't connect to broker. Error:", err.Error())
	} else {
		log.Info("--------------Connected----------------")
	}
	defer mqtt.Stop()

	responder := discovery.NewServiceDiscoveryResponder(mqtt)
	responder.RegisterResource(model.GetDiscoveryResource())
	responder.Start()

	fimpHandler := handler.NewFimpTibberHandler(mqtt, configs.WorkDir)
	fimpHandler.Start()

	log.Info("--------------Started handler----------")

	mqtt.Subscribe("pt:j1/mt:cmd/rt:ad/rn:tibber/ad:1")
	mqtt.Subscribe("pt:j1/mt:cmd/rt:dev/rn:tibber/ad:1/#")
	// Listen for the factory reset event
	mqtt.Subscribe("pt:j1/mt:evt/rt:ad/rn:gateway/ad:1")

	log.Info("Subscribing to topic: pt:j1/mt:cmd/rt:ad/rn:tibber/ad:1")
	log.Info("Subscribing to topic: pt:j1/mt:cmd/rt:dev/rn:tibber/ad:1/#")

	select {}
}
