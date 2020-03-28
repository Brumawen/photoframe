package main

import (
	"testing"
)

func TestCanGetBingImages(t *testing.T) {
	c := Config{
		Provider:   0,
		ImgCount:   8,
		Resolution: 0,
	}
	c.SetDefaults()

	i := IodBing{Config: c}
	l, err := i.GetImages()
	if err != nil {
		t.Error(err)
	}
	t.Log("Number of images = ", len(l))
	if len(l) != 8 {
		t.Error("Only", len(l), "images returned, expected 8.")
	}
}
