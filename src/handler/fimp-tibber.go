package handler

import (
	"github.com/futurehomeno/fimpgo"
	scribble "github.com/nanobox-io/golang-scribble"
	log "github.com/sirupsen/logrus"
	"github.com/tskaard/fh-tibber/model"
	tibber "github.com/tskaard/tibber-golang"
)

// FimpTibberHandler structure
type FimpTibberHandler struct {
	inboundMsgCh fimpgo.MessageCh
	mqt          *fimpgo.MqttTransport
	db           *scribble.Driver
	state        model.State
	tibber       *tibber.Client
	streams      map[string]*tibber.Stream
	tibberMsgCh  tibber.MsgChan
}

// NewFimpTibberHandler construct new handler
func NewFimpTibberHandler(transport *fimpgo.MqttTransport, stateFile string) *FimpTibberHandler {
	t := &FimpTibberHandler{
		inboundMsgCh: make(fimpgo.MessageCh, 5),
		mqt:          transport,
		tibber:       tibber.NewClient(""),
		streams:      make(map[string]*tibber.Stream),
		tibberMsgCh:  make(tibber.MsgChan),
		state:        model.State{},
	}
	t.db, _ = scribble.New(stateFile, nil)
	t.mqt.RegisterChannel("ch1", t.inboundMsgCh)
	return t
}

// Start handler
func (t *FimpTibberHandler) Start() error {
	if err := t.db.Read("data", "state", &t.state); err != nil {
		log.Info("Error loading state from file: ", err)
		t.state.Connected = false
		log.Debug("setting state connected to false")
		if err = t.db.Write("data", "state", t.state); err != nil {
			log.Error("Did not manage to write to file: ", err)
		}
	}
	t.tibber.Token = t.state.AccessToken
	if t.state.Connected {
		for _, home := range t.state.Homes {
			stream := tibber.NewStream(home.ID, t.tibber.Token)
			stream.StartSubscription(t.tibberMsgCh)
			t.streams[home.ID] = stream
		}
	}
	var errr error
	go func(msgChan fimpgo.MessageCh) {
		for {
			select {
			case newMsg := <-msgChan:
				t.routeFimpMessage(newMsg)
			}
		}
	}(t.inboundMsgCh)

	go func(msgChan tibber.MsgChan) {
		for {
			select {
			case msg := <-msgChan:
				t.routeTibberMessage(msg)
			}
		}
	}(t.tibberMsgCh)
	return errr
}

func (t *FimpTibberHandler) routeTibberMessage(msg *tibber.StreamMsg) {
	//log.Debug("New tibber msg")
	for _, home := range t.state.Homes {
		if home.ID == msg.HomeID {
			t.sendPowerMsg(msg.HomeID, float64(msg.Payload.Data.LiveMeasurement.Power), nil)
		}
	}
}

func (t *FimpTibberHandler) routeFimpMessage(newMsg *fimpgo.Message) {
	log.Debug("New fimp msg")
	switch newMsg.Payload.Type {
	case "cmd.system.disconnect":
		t.systemDisconnect(newMsg)

	case "cmd.system.get_connect_params":
		t.systemGetConnectionParameter(newMsg)

	case "cmd.system.connect":
		t.systemConnect(newMsg)

	case "cmd.system.sync":
		t.systemSync(newMsg)

	case "cmd.sensor.get_report":
		log.Debug("cmd.sensor.get_report")
		if newMsg.Payload.Service != "sensor_power" {
			log.Error("sensor.get_report - Wrong service")
			break
		}
		log.Debug("Maby do something here later")
	}

}
