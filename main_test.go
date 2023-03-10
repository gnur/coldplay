package main

import (
	"testing"
	"time"
)

func Test_justStarted(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name   string
		points []Measurement
		want   bool
	}{
		{
			name: "no movement",
			points: []Measurement{
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := justStarted(tt.points); got != tt.want {
				t.Errorf("justStarted() = %v, want %v", got, tt.want)
			}
		})
	}
}
