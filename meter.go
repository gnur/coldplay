package main

import (
	"fmt"

	"github.com/d2r2/go-i2c"
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
	dev *i2c.I2C
}

func NewMeter() (*meter, error) {

	fmt.Println("ok")
	dev, err := i2c.NewI2C(TFLUNA_ADDR, 1)
	if err != nil {
		return nil, err
	}
	m := meter{
		dev: dev,
	}

	return &m, nil
}
