package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"

	"github.com/stianeikeland/go-rpio/v4"
)

func main() {

	dist := read()
	fmt.Println("Distance:", dist)
}

func print(s string) {
	fmt.Println(time.Now().UnixMicro(), s)
}


func writeToInflux() {
	token := os.Getenv("INFLUXDB_TOKEN")
	url := "http://uranus:8086"
	client := influxdb2.NewClient(url, token)

	org := "gnur"
	bucket := "coldplay"
	writeAPI := client.WriteAPIBlocking(org, bucket)
	for value := 0; value < 5; value++ {
		tags := map[string]string{
			"tagname1": "tagvalue1",
		}
		fields := map[string]interface{}{
			"field1": value,
		}
		point := write.NewPoint("measurement1", tags, fields, time.Now())
		fmt.Println("Writing: %i", value)

		if err := writeAPI.WritePoint(context.Background(), point); err != nil {
			log.Fatal(err)
		}
		time.Sleep(1 * time.Second) // separate points by 1 second
	}
}
