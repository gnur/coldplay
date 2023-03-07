package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
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

	return

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

func (cold *coldplay) brain() {
	for point := range cold.meter.ch {
		fmt.Println(point)
	}
}
