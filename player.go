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
	ctrl   *beep.Ctrl
	volume float64
	music  *effects.Volume
	floors []*effects.Volume
}

func newPlayer() (*player, error) {
	f, err := os.Open("/home/erwin/maxpayne.mp3")
	if err != nil {

		return nil, fmt.Errorf("Could not open mp3: %w", err)
	}
	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	defer streamer.Close()
	if err != nil {
		return nil, fmt.Errorf("Could not decode mp3: %w", err)
	}
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)

	ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, buffer.Streamer(0, buffer.Len())), Paused: true}

	musicLoop := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   3,
		Silent:   false,
	}

	speaker.Play(musicLoop)

	var floors []*effects.Volume

	for i := 0; i <= 2; i++ {
		f, err := os.Open(fmt.Sprintf("/home/erwin/etage%v.mp3", i))
		if err != nil {
			return nil, fmt.Errorf("Could not open mp3: %w", err)
		}
		defer f.Close()

		streamer, _, err := mp3.Decode(f)
		if err != nil {
			return nil, fmt.Errorf("Could not decode mp3: %w", err)
		}
		defer streamer.Close()

		buffer := beep.NewBuffer(format)
		buffer.Append(streamer)

		ctrl := &beep.Ctrl{Streamer: buffer.Streamer(0, buffer.Len()), Paused: false}
		floors = append(floors, &effects.Volume{
			Streamer: ctrl,
			Base:     2,
			Volume:   3,
			Silent:   false,
		})
	}

	return &player{
		ctrl:   ctrl,
		music:  musicLoop,
		floors: floors,
	}, nil
}

func (p *player) playAnnouncement(floor int) {
	p.floors[floor].Volume = p.volume + 2
	speaker.Play(p.floors[floor])
}

func (p *player) setVolume(f float64) {
	if f > 3 || f < 0 {
		return
	}
	p.volume = f
	speaker.Lock()
	p.music.Volume = f
	speaker.Unlock()
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
