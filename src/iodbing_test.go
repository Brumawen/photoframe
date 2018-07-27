package main

import (
	"testing"
)

func TestCanGetImages(t *testing.T) {
	c := Config{
		Resolution: 0,
		Provider:   0,
		ImgCount:   8,
	}

	i := IodBing{Config: c}
	l, err := i.GetImages()
	if err != nil {
		t.Error(err)
	}
	if len(l) != 8 {
		t.Error("Only", len(l), "images returned, expected 8.")
	}
}
