package main

import (
	"fmt"
	"testing"
)

func TestCanGetWeatherForecast(t *testing.T) {
	c := Config{}
	c.ReadFromFile("config.json")

	w, err := GetForecast(c)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(w)
}
