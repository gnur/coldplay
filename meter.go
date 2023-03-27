package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type meter struct {
	sync.Mutex
	ch chan Measurement
}

func NewMeter() (*meter, error) {

	met := meter{
		ch: make(chan Measurement),
	}

	nc, err := nats.Connect("uranus")
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to nats: %w", err)
	}

	loc, err := time.LoadLocation("Local")
	if err != nil {
		return nil, fmt.Errorf("Failed to load local timezone: %w", err)
	}

	nc.Subscribe("coldplay.measurement", func(m *nats.Msg) {
		var point Measurement
		err := json.Unmarshal(m.Data, &point)
		if err != nil {
			fmt.Println(err)
			return
		}

		logrus.WithFields(logrus.Fields{
			"height":      point.Height,
			"temperature": point.Temperature,
		}).Info("Got measurement")
		point.Timestamp = point.Timestamp.In(loc)
		met.ch <- point

	})
	//noop for now
	//will read nats later

	return &met, nil
}
