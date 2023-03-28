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

const (
	GROUND_FLOOR_HEIGHT = 2.0
	MIDDLE_FLOOR_HEIGHT = 278.5
	TOP_FLOOR_HEIGHT    = 544
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

type scientist struct {
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

	cold := scientist{
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

func (science *scientist) brain() {
	history := []Measurement{}
	lastWrite := time.Now().Add(-5 * time.Minute)

	for point := range science.meter.ch {

		var b bytes.Buffer
		science.tpl.ExecuteTemplate(&b, "height", V{
			Height: point.Height,
			Temp:   point.Temperature,
		})

		science.sse.Publish("measurements", &sse.Event{
			Data: b.Bytes(),
		})
		history = append(history, point)

		if len(history) > 10 {
			history = history[1:]
		}

		if isMoving(history) {
			science.ll.WithFields(logrus.Fields{
				"height":      point.Height,
				"temperature": point.Temperature,
			}).Info("Writing to prometheus because we're moving")
			lastWrite = time.Now()
			science.writer.ch <- point
		} else if time.Since(lastWrite) > 10*time.Second {
			science.ll.WithFields(logrus.Fields{
				"height":      point.Height,
				"temperature": point.Temperature,
			}).Info("Writing to prometheus because it's been too long")
			lastWrite = time.Now()
			science.writer.ch <- point
		}

		if isMoving(history) {
			vol := 3 - (3 * point.Height / TOP_FLOOR_HEIGHT)
			science.player.setVolume(vol)
		}

		if justChangedMovement(history) {
			science.ll.Info("movement changed")

			if isMoving(history) {
				science.ll.Info("Starting music")
				science.player.start()
			} else {

				if at, floor := isAtFloor(history); at {
					science.ll.WithField("floor", floor).Info("stopping music and playing floor announcement")
					science.player.stop()
					science.player.playAnnouncement(floor)
				} else {
					science.ll.Info("not stopping music because we're between floors")
				}
			}
		}

	}
}

func isAtFloor(points []Measurement) (bool, int) {
	cur := points[len(points)-1].Height

	if math.Abs(cur-GROUND_FLOOR_HEIGHT) < 6 {
		return true, 0
	}

	if math.Abs(cur-MIDDLE_FLOOR_HEIGHT) < 6 {
		return true, 1
	}

	if math.Abs(cur-TOP_FLOOR_HEIGHT) < 6 {
		return true, 2
	}

	return false, -1
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
