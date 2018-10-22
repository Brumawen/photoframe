package main

import (
	"fmt"
	"testing"
)

func TestCanGetNatGeoImages(t *testing.T) {
	//t.Errorf("Wotcher")
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
		t.Error("Only", len(l), "images returned, expected 12.")
	}
	for _, x := range l {
		fmt.Println("Image", x.Name, x.ImagePath)
	}
}
