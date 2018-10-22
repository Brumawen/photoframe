package main

import (
	"fmt"
	"testing"
)

func TestCanGetWeatherForecast(t *testing.T) {
	w, err := GetForecast()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(w)
}
