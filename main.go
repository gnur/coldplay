package main

import (
	"bytes"
	_ "embed"
	"html/template"
	"math"
	"net/http"
	"strings"
	"time"

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
	Timestamp   time.Time
}

type coldplay struct {
	meter  *meter
	writer *writer
	player *player
	ll     *logrus.Entry
	sse    *sse.Server
	tpl    *template.Template
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

	logrus.Info("Loading player")
	p, err := newPlayer()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create player")
	}
	logrus.Info("Player loaded, starting")

	server := sse.New()
	server.AutoReplay = false
	server.CreateStream("measurements")
	server.CreateStream("writes")
	server.CreateStream("playerupdates")

	cold := coldplay{
		meter:  m,
		writer: w,
		player: p,
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
			Height: 0,
			Temp:   0,
		})
	})

	http.ListenAndServe(":10211", mux)
}

func (cold *coldplay) brain() {
	history := []Measurement{}
	lastWrite := time.Now().Add(-5 * time.Minute)

	for point := range cold.meter.ch {

		var b bytes.Buffer
		cold.tpl.ExecuteTemplate(&b, "height", V{
			Height: point.Height,
			Temp:   point.Temperature,
		})

		cold.sse.Publish("measurements", &sse.Event{
			Data: b.Bytes(),
		})
		history = append(history, point)

		if len(history) > 10 {
			history = history[1:]
		}

		if isMoving(history) {
			cold.ll.WithFields(logrus.Fields{
				"height":      point.Height,
				"temperature": point.Temperature,
			}).Info("Writing to prometheus because we're moving")
			lastWrite = time.Now()
			cold.writer.ch <- point
		} else if time.Since(lastWrite) > 30*time.Second {
			cold.ll.WithFields(logrus.Fields{
				"height":      point.Height,
				"temperature": point.Temperature,
			}).Info("Writing to prometheus because it's been too long")
			lastWrite = time.Now()
			cold.writer.ch <- point
		}

		if isMoving(history) {
			vol := 3 - (3 * point.Height / TOP_FLOOR_HEIGHT)
			cold.player.setVolume(vol)
		}

		if justChangedMovement(history) {
			cold.ll.Info("movement changed")

			if isMoving(history) {
				cold.ll.Info("Starting music")
				cold.player.start()
			} else {

				if !isBetweenFloors(history) {
					cold.ll.Info("stopping music")
					cold.player.stop()
				} else {
					cold.ll.Info("not stopping music because we're between floors")
				}
			}
		}

		if isStale(history) {
			cold.ll.Error("Resetting device because measurements have become stale")
			cold.meter.reset()
		}
	}
}

func isBetweenFloors(points []Measurement) bool {
	cur := points[len(points)-1].Height

	if math.Abs(cur-GROUND_FLOOR_HEIGHT) < 5 {
		return false
	}

	if math.Abs(cur-MIDDLE_FLOOR_HEIGHT) < 5 {
		return false
	}

	if math.Abs(cur-TOP_FLOOR_HEIGHT) < 5 {
		return false
	}

	return true
}

func justChangedMovement(points []Measurement) bool {
	n := len(points)
	if n < 3 {
		//not enough data to check for movement
		return false
	}

	oldCheck := points[0 : n-1]
	newCHeck := points[n-2 : n]

	old := isMoving(oldCheck)
	cur := isMoving(newCHeck)

	return cur != old
}

func isMoving(points []Measurement) bool {
	if len(points) < 2 {
		//not enough data to check for movement
		return false
	}

	a := points[len(points)-1]
	b := points[len(points)-2]

	distance := a.Height - b.Height
	time := a.Timestamp.Sub(b.Timestamp).Seconds()

	speed := distance / time

	if math.Abs(speed) > 5 {
		// moving
		return true
	}

	return false
}

func isStale(points []Measurement) bool {
	if len(points) < 5 {
		//not enough data to check for movement
		return false
	}

	last := points[0].Height
	for _, p := range points {
		if last != p.Height {
			//any change is okay to prevent staleness
			return false
		}
	}

	return true
}
