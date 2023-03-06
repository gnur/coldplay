package main

import (
	"fmt"
	"time"

	"github.com/d2r2/go-i2c"
	"github.com/d2r2/go-logger"
)

const (
	TFLUNA_ADDR  = 0x10
	MODE_ADDR    = 0x23
	TRIGGER_MODE = 0x01
	TRIGGER_ADDR = 0x24

	GROUND_FLOOR_HEIGHT = 2.0
	MIDDLE_FLOOR_HEIGHT = 276.5
	TOP_FLOOR_HEIGHT    = 544

	SAMPLES = 50
)

type meter struct {
	samples int
	dev     *i2c.I2C
}

func NewMeter() (*meter, error) {

	fmt.Println("ok")
	dev, err := i2c.NewI2C(TFLUNA_ADDR, 1)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize i2c dev: %w", err)
	}
	m := meter{
		dev:     dev,
		samples: SAMPLES,
	}
	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)

	d, t, err := m.measure()
	if err != nil {
		return nil, fmt.Errorf("Failed to do init measure: %w", err)
	}
	fmt.Printf("Distance: %v, temp: %v\n", d, t)

	go m.measureLoop()

	return &m, nil
}

func (m *meter) measureLoop() {
	for {

		var dSum uint = 0
		var tSum uint = 0
		samples := 0

		for i := 0; i < m.samples; i++ {
			time.Sleep(10 * time.Millisecond)
			d, t, err := m.measure()
			if err != nil {
				continue
			}
			samples++
			dSum += d
			tSum += t
		}

		if samples != m.samples {
			//TODO: add warning
			continue
		}

		distance := float64(dSum) / float64(samples)
		temp := float64(tSum) / float64(samples)

		fmt.Println("Distance: ", distance)
		fmt.Println("Temperat: ", temp)

	}
}

func (m *meter) measure() (uint, uint, error) {

	err := m.dev.WriteRegU8(TRIGGER_ADDR, TRIGGER_MODE)
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to write: %w", err)
	}
	time.Sleep(10 * time.Millisecond)

	buf, total, err := m.dev.ReadRegBytes(0x00, 6)
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to read: %w", err)
	}
	if total != 6 {
		return 0, 0, fmt.Errorf("Not enough bytes read, expected 6, got: %v", total)
	}

	distance := 256*uint(buf[1]) + uint(buf[0])
	temp := 256*uint(buf[5]) + uint(buf[4])

	return distance, temp, nil
}
