package main

import (
	"testing"
)

func TestCanGetPexelsImages(t *testing.T) {
	c := Config{
		Provider:   2,
		ImgCount:   8,
		Resolution: 0,
	}
	c.SetDefaults()

	i := Pexels{Config: c}
	l, err := i.GetImages()
	if err != nil {
		t.Error(err)
	}
	if len(l) != 8 {
		t.Error("Only", len(l), "images returned, expected 8.")
	}
}
