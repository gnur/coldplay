package main

import (
	"fmt"
	"log"
	"time"
)

type Measurement struct {
	Height      float64
	Temperature float64
	Timestamp   time.Time
}

type coldplay struct {
	meter  *meter
	writer *writer
}

func main() {

	fmt.Println("Starting")

	m, err := NewMeter()
	if err != nil {
		log.Fatal(err)
	}

	w, err := NewWriter()
	if err != nil {
		log.Fatal(err)
	}

	cold := coldplay{
		meter:  m,
		writer: w,
	}

	go cold.brain()

	select {}

}

func (cold *coldplay) brain() {
	brain := []Measurement{}
	wasMoving := false

	for point := range cold.meter.ch {
		brain = append(brain, point)

		if len(brain) > 10 {
			brain = brain[1:]
		}

		if isMoving(brain) && !wasMoving {

			//something changed
			//music play or music pause
		}

	}
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

	if speed > 5 {
		// moving
		return true
	}

	fmt.Println(distance, time)

	//do some fancy calculation to determine current speed

	return false
}
