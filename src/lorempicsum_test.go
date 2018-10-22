package main

import (
	"testing"
)

func TestCanGetLoremImages(t *testing.T) {
	c := Config{
		Provider:   0,
		ImgCount:   8,
		Resolution: 0,
	}
	c.SetDefaults()

	i := LoremPicsum{Config: c}
	l, err := i.GetImages()
	if err != nil {
		t.Error(err)
	}
	if len(l) != 8 {
		t.Error("Only", len(l), "images returned, expected 8.")
	}
}
