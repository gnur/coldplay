package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/d2r2/go-i2c"
	"github.com/d2r2/go-logger"
)

const (
	TFLUNA_ADDR = 0x10

	MODE_ADDR    = 0x23
	TRIGGER_MODE = 0x01

	RESET_ADDR  = 0x21
	RESET_BYTES = 0x02

	TRIGGER_BYTES = 0x01
	TRIGGER_ADDR  = 0x24

	GROUND_FLOOR_HEIGHT = 2.0
	MIDDLE_FLOOR_HEIGHT = 276.5
	TOP_FLOOR_HEIGHT    = 544

	OFFSET = 544.3

	SAMPLES  = 20
	MAX_TEMP = 33
)

type meter struct {
	sync.Mutex
	samples int
	dev     *i2c.I2C
	ch      chan Measurement
}

func NewMeter() (*meter, error) {

	dev, err := i2c.NewI2C(TFLUNA_ADDR, 1)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize i2c dev: %w", err)
	}
	m := meter{
		dev:     dev,
		samples: SAMPLES,
		ch:      make(chan Measurement),
	}
	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)

	// reset will make sure we start with a clean slate
	m.reset()

	_, _, err = m.measure()
	if err != nil {
		return nil, fmt.Errorf("Failed to do init measure: %w", err)
	}

	go m.measureLoop()

	return &m, nil
}

func (m *meter) measureLoop() {
	for {
		time.Sleep(100 * time.Millisecond)

		var dSum uint = 0
		var tSum uint = 0
		samples := 0

		for i := 0; i < m.samples; i++ {
			time.Sleep(25 * time.Millisecond)
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
			fmt.Println("Samples were dropped, skipping")
			continue
		}

		distance := float64(dSum) / float64(samples)
		temp := float64(tSum) / float64(samples)

		m.ch <- Measurement{
			Temperature: temp / 100.0,
			Height:      OFFSET - distance,
			Timestamp:   time.Now(),
		}

		//safeguard to prevent heating up if it is over 33 degrees
		if temp > 3300 {
			fmt.Println("Sleeping for a while to prevent overheating")
			time.Sleep(20 * time.Second)
		}
	}
}

func (m *meter) measure() (uint, uint, error) {

	m.Lock()
	err := m.dev.WriteRegU8(TRIGGER_ADDR, TRIGGER_BYTES)
	m.Unlock()
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to write: %w", err)
	}
	time.Sleep(10 * time.Millisecond)

	m.Lock()
	buf, total, err := m.dev.ReadRegBytes(0x00, 6)
	m.Unlock()
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

func (m *meter) reset() error {

	// first reset the device
	m.Lock()
	err := m.dev.WriteRegU8(RESET_ADDR, RESET_BYTES)
	m.Unlock()
	if err != nil {
		return fmt.Errorf("Failed to setup trigger mode: %w", err)
	}

	//then make sure we set it up in trigger mode again
	m.Lock()
	err = m.dev.WriteRegU8(MODE_ADDR, TRIGGER_MODE)
	m.Unlock()
	if err != nil {
		return fmt.Errorf("Failed to setup trigger mode: %w", err)
	}
	return nil
}
