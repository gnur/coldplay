package main

import (
	"context"
	"fmt"
	"time"

	"github.com/castai/promwrite"
)

type writer struct {
	ch   chan Measurement
	prom *promwrite.Client
}

func NewWriter() (*writer, error) {

	return &writer{
		ch:   make(chan Measurement),
		prom: promwrite.NewClient("http://uranus:9090/api/v1/write"),
	}, nil

}

func (w *writer) remoteWriter() {

	loc, err := time.LoadLocation("Local")
	if err != nil {
		fmt.Println(err)
		return
	}

	for point := range w.ch {

		fmt.Printf("Received a measurement\nheight: %f\ntemperature: %f\n", point.Height, point.Temperature)
		point.Timestamp = point.Timestamp.In(loc)

		_, err = w.prom.Write(context.Background(), &promwrite.WriteRequest{
			TimeSeries: []promwrite.TimeSeries{
				{
					Labels: []promwrite.Label{
						{
							Name:  "__name__",
							Value: "object_height",
						},
						{
							Name:  "device",
							Value: "elevator",
						},
					},
					Sample: promwrite.Sample{
						Time:  point.Timestamp,
						Value: point.Height,
					},
				},
				{
					Labels: []promwrite.Label{
						{
							Name:  "__name__",
							Value: "object_temperature",
						},
						{
							Name:  "device",
							Value: "luna",
						},
					},
					Sample: promwrite.Sample{
						Time:  point.Timestamp,
						Value: point.Temperature,
					},
				},
			},
		})
		if err != nil {
			fmt.Println(err)
		}
	}

}
