package main

import (
	"testing"
)

func TestCanGetNatGeoImages(t *testing.T) {
	c := Config{
		Provider:   2,
		ImgCount:   12,
		Resolution: 0,
	}
	c.SetDefaults()

	i := NatGeo{Config: c}
	l, err := i.GetImages()
	if err != nil {
		t.Error(err)
	}
	if len(l) != c.ImgCount {
		t.Error("Only", len(l), "images returned, expected 8.")
	}
}
