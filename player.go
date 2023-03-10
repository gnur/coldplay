package main

import (
	"fmt"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type player struct {
	ctrl *beep.Ctrl
}

func newPlayer() (*player, error) {
	f, err := os.Open("/home/erwin/maxpayne.mp3")
	if err != nil {

		return nil, fmt.Errorf("Could not open mp3: %w", err)
	}

	streamer, format, err := mp3.Decode(f)
	defer streamer.Close()

	if err != nil {
		return nil, fmt.Errorf("Could not decode mp3: %w", err)
	}
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)

	ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, buffer.Streamer(0, buffer.Len())), Paused: true}

	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   3,
		Silent:   false,
	}

	speaker.Play(volume)

	return &player{
		ctrl: ctrl,
	}, nil
}

func (p *player) start() {
	speaker.Lock()
	p.ctrl.Paused = false
	speaker.Unlock()
}

func (p *player) stop() {
	speaker.Lock()
	p.ctrl.Paused = true
	speaker.Unlock()
}
