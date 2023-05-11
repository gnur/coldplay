package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"go.bug.st/serial"
)

type meter struct {
	sync.Mutex
	ch chan Measurement
}

func NewMeter() (*meter, error) {

	met := meter{
		ch: make(chan Measurement),
	}

	mode := serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open("/dev/ttyAMA0", &mode)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Starting read")

	sample_rate_packet := []byte{0x5A, 0x06, 0x03, 0x01, 0x00, 0x00}
	_, err = port.Write(sample_rate_packet)
	if err != nil {
		return nil, fmt.Errorf("Unable to set sample rate: %w", err)
	}

	output_format_packet := []byte{0x5A, 0x05, 0x05, 0x06, 0x00}
	_, err = port.Write(output_format_packet)
	if err != nil {
		return nil, fmt.Errorf("Unable to set output format: %w", err)
	}

	loc, err := time.LoadLocation("Local")
	if err != nil {
		return nil, fmt.Errorf("Failed to load local timezone: %w", err)
	}

	buff := make([]byte, 9)

	var distance float64
	var temperature float64
	var strength float64
	var point Measurement

	go func() {
		for {
			n, err := port.Read(buff)
			if err != nil {
				return
			}
			if n == 0 {
				fmt.Println("EOF")
				break
			}
			if buff[0] != 0x59 || buff[1] != 0x59 {
				//invalid signature
				fmt.Println("invalid signature")
				continue
			}

			distance = float64(buff[2]) + float64(buff[3])*256.0
			distance = distance / 10 //from mm to cm

			strength = float64(buff[4]) + float64(buff[5])*256.0

			temperature = float64(buff[6]) + float64(buff[7])*256.0
			temperature = temperature/8 - 256

			point = Measurement{
				Height:      distance,
				Temperature: temperature,
				Strength:    strength,
				Timestamp:   time.Now().In(loc),
			}
			met.ch <- point
		}
	}()

	return &met, nil
}
