package main

import (
	"testing"
)

func TestCanGetFileFolderImages(t *testing.T) {
	c := Config{}
	c.SetDefaults()

	i := FileFolder{Config: c}
	l, err := i.GetImages()
	if err != nil {
		t.Error(err)
	}
	t.Log("Number of images = ", len(l))
	if len(l) != 22 {
		t.Error("Only", len(l), "image returned, exprected 22.")
	}
}
