package main

import (
	_ "embed"
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
			name: "1 measurement",
			points: []Measurement{
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now,
				},
			},
			want: false,
		},
		{
			name: "no movement",
			points: []Measurement{
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now,
				},
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now.Add(1 * time.Second),
				},
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now.Add(2 * time.Second),
				},
			},
			want: false,
		},
		{
			name: "just started moving",
			points: []Measurement{
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now,
				},
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now.Add(1 * time.Second),
				},
				{
					Height:      10,
					Temperature: 0,
					Timestamp:   now.Add(2 * time.Second),
				},
			},
			want: true,
		},
		{
			name: "just stopped moving",
			points: []Measurement{
				{
					Height:      10,
					Temperature: 0,
					Timestamp:   now,
				},
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now.Add(1 * time.Second),
				},
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now.Add(2 * time.Second),
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := justChangedMovement(tt.points); got != tt.want {
				t.Errorf("justStarted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isMoving(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name   string
		points []Measurement
		want   bool
	}{
		{
			name:   "no measurement",
			points: []Measurement{},
			want:   false,
		},
		{
			name: "1 measurement",
			points: []Measurement{
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now,
				},
			},
			want: false,
		},
		{
			name: "no movement",
			points: []Measurement{
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now,
				},
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now.Add(1 * time.Second),
				},
			},
			want: false,
		},
		{
			name: "slight movement",
			points: []Measurement{
				{
					Height:      1,
					Temperature: 0,
					Timestamp:   now,
				},
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now.Add(1 * time.Second),
				},
			},
			want: false,
		},
		{
			name: "big movement",
			points: []Measurement{
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now,
				},
				{
					Height:      10,
					Temperature: 0,
					Timestamp:   now.Add(1 * time.Second),
				},
			},
			want: true,
		},
		{
			name: "big movement down",
			points: []Measurement{
				{
					Height:      10,
					Temperature: 0,
					Timestamp:   now,
				},
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now.Add(1 * time.Second),
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMoving(tt.points); got != tt.want {
				t.Errorf("isMoving() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isStale(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name   string
		points []Measurement
		want   bool
	}{
		{
			name: "too few movements",
			points: []Measurement{
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now,
				},
				{
					Height:      0,
					Temperature: 0,
					Timestamp:   now.Add(1 * time.Second),
				},
			},
			want: false,
		},
		{
			name: "slight movement",
			points: []Measurement{
				{
					Height: 1,
				},
				{
					Height: 0,
				},
				{
					Height: 1,
				},
				{
					Height: 1,
				},
				{
					Height: 1,
				},
				{
					Height: 1,
				},
			},
			want: false,
		},
		{
			name: "stale",
			points: []Measurement{
				{
					Height: 1,
				},
				{
					Height: 1,
				},
				{
					Height: 1,
				},
				{
					Height: 1,
				},
				{
					Height: 1,
				},
				{
					Height: 1,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isStale(tt.points); got != tt.want {
				t.Errorf("isStale() = %v, want %v", got, tt.want)
			}
		})
	}
}
