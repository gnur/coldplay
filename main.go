package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/castai/promwrite"
	"github.com/gnur/beep/mp3"
	"github.com/gnur/beep/speaker"
	"github.com/nats-io/nats.go"
)

type Measurement struct {
	Height      float64
	Temperature float64
	Timestamp   time.Time
}

func main() {
	f, err := os.Open("music.mp3")
	if err != nil {

		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	defer streamer.Close()

	if err != nil {
		log.Fatal(err)
	}
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(streamer)
	select {}

}

func test() {

	cl := promwrite.NewClient("http://uranus:9090/api/v1/write")

	nc, err := nats.Connect("uranus")
	if err != nil {
		fmt.Println(err)
		return
	}

	loc, err := time.LoadLocation("Local")
	if err != nil {
		fmt.Println(err)
		return
	}

	nc.Subscribe("coldplay.measurement", func(m *nats.Msg) {
		var point Measurement
		err := json.Unmarshal(m.Data, &point)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Received a measurement\nheight: %f\ntemperature: %f\n", point.Height, point.Temperature)
		point.Timestamp = point.Timestamp.In(loc)

		_, err = cl.Write(context.Background(), &promwrite.WriteRequest{
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

	})

	wait := make(chan bool)
	<-wait

}
