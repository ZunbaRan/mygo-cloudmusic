package app

import (
	"testing"
	"time"
)

func TestClockAndPlays(t *testing.T) {
	if s := clock(185 * time.Second); s != "3:05" {
		t.Error(s)
	}
	if s := plays(123456789); s != "123.5M" {
		t.Error(s)
	}
	if s := plays(56789); s != "57K" {
		t.Error(s)
	}
}
