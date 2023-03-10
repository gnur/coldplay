package main

import (
	"log"
	"os"
	"time"

	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func play() {
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
