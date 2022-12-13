package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type Measurement struct {
	Height    float64
	Timestamp time.Time
}

func main() {
	token := os.Getenv("INFLUXDB_TOKEN")
	url := "http://uranus:8086"
	client := influxdb2.NewClient(url, token)

	org := "gnur"
	bucket := "coldplay"
	writeAPI := client.WriteAPIBlocking(org, bucket)

	nc, err := nats.Connect("uranus")
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

		fmt.Printf("Received a measurement: %f \n", point.Height)

		tags := map[string]string{
			"object": "elevator",
		}
		fields := map[string]interface{}{
			"height": point.Height,
		}

		influxPoint := write.NewPoint("measurement1", tags, fields, point.Timestamp)

		if err := writeAPI.WritePoint(context.Background(), influxPoint); err != nil {
			fmt.Println(err)
			return
		}
	})

	wait := make(chan bool)
	<-wait

}
