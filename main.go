package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/r3labs/sse/v2"
	"github.com/sirupsen/logrus"
)

type V struct {
	Height float64
	Temp   float64
}

//go:embed html/sse.js
var ssejs []byte

//go:embed html/htmx.min.js
var htmxminjs []byte

//go:embed html/bulma.min.css
var bulmamincss []byte

//go:embed html/templates.html
var templates string

type Measurement struct {
	Height      float64
	Temperature float64
	Strength    float64
	Timestamp   time.Time
}

type scientist struct {
	meter      *meter
	writer     *writer
	ll         *logrus.Entry
	sse        *sse.Server
	tpl        *template.Template
	lastHeight float64
	lastTemp   float64
}

func main() {

	logrus.Info("Starting")

	tpl := template.New("")
	tpl.Funcs(templateFunctions)
	tpl, err := tpl.Parse(strings.Replace(templates, "\n", "", -1))
	if err != nil {
		logrus.WithError(err).Fatal("could not parse templates")
	}

	m, err := NewMeter()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create meter")
	}

	w, err := NewWriter()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create writer")
	}

	server := sse.New()
	server.AutoReplay = false
	server.CreateStream("measurements")
	server.CreateStream("writes")
	server.CreateStream("playerupdates")

	cold := scientist{
		meter:  m,
		writer: w,
		ll:     logrus.WithField("app", "coldplay"),
		sse:    server,
		tpl:    tpl,
	}

	go cold.brain()

	// Create a new Mux and set the handler
	mux := http.NewServeMux()
	mux.HandleFunc("/events", server.ServeHTTP)
	mux.HandleFunc("/sse.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/javascript")
		w.Write(ssejs)
	})
	mux.HandleFunc("/htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/javascript")
		w.Write(htmxminjs)
	})
	mux.HandleFunc("/bulma.min.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/css")
		w.Write(bulmamincss)
	})

	mux.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		tpl.ExecuteTemplate(w, "index.html", V{
			Height: cold.lastHeight,
			Temp:   cold.lastTemp,
		})
	})

	http.ListenAndServe(":10211", mux)
}

func (science *scientist) brain() {

	counter := 0

	for point := range science.meter.ch {

		science.lastHeight = point.Height
		science.lastTemp = point.Height

		var b bytes.Buffer
		science.tpl.ExecuteTemplate(&b, "height", V{
			Height: point.Height,
			Temp:   point.Temperature,
		})

		science.sse.Publish("measurements", &sse.Event{
			Data: b.Bytes(),
		})

		science.writer.ch <- point

		err := influxWrite(context.Background(), point.Height)
		if err != nil {
			fmt.Println(err)
		}

		counter++
		if counter%10 == 0 {
			fmt.Println("Processed another 10 measurements")
		}

	}
}

func influxWrite(ctx context.Context, h float64) error {
	// Create write client
	url := "http://uranus:8086"
	token := ""

	writeClient := influxdb2.NewClientWithOptions(url, token, influxdb2.DefaultOptions().SetFlushInterval(30000))

	// Define write API
	org := "gnur"
	bucket := "Project Coldplay"
	writeAPI := writeClient.WriteAPI(org, bucket)

	point := influxdb2.NewPointWithMeasurement("coldplay").
		AddTag("device", "elevator").
		AddField("object_height", h)

	writeAPI.WritePoint(point)

	return nil
}
