package main

import (
	"math"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

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
}

func main() {

	logrus.Info("Starting")

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

	cold := coldplay{
		meter:  m,
		writer: w,
		player: p,
		ll:     logrus.WithField("app", "coldplay"),
	}

	go cold.brain()

	// listening on a port without anything just to make it exclusive
	http.ListenAndServe(":10211", nil)
}

func (cold *coldplay) brain() {
	history := []Measurement{}
	lastWrite := time.Now().Add(-5 * time.Minute)

	for point := range cold.meter.ch {
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
