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

func read() int64 {
	err := rpio.Open()
	if err != nil {
		fmt.Println(err)
	}
	trig := rpio.Pin(8)
	echo := rpio.Pin(8)

	trig.Output()
	trig.Low()

	echo.Input()

	start := time.Now()
	stop := time.Now()

	print("trigger")
	trig.High()
	time.Sleep(5000 * time.Microsecond)
	trig.Low()

	echo.PullDown()
	echo.Detect(rpio.RiseEdge)

	//I think this is blocking ?
	echo.EdgeDetected()
	echo.Detect(rpio.NoEdge)

	print("done")
	stop = time.Now()

	elapsed := stop.Sub(start)

	distance := (elapsed.Milliseconds() * 34300) / 2

	rpio.Close()

	return distance
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
