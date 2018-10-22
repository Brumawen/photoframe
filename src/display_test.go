package main

import "testing"

func TestCanBuildDisplayImages(t *testing.T) {
	c := Config{}
	c.ReadFromFile("config.json")
	c.SetDefaults()
	s := Server{Config: &c}
	d := Display{Srv: &s}

	d.Run()
	if d.LastErr != nil {
		t.Error(d.LastErr)
	}
}
