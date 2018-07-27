package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

// Forecast holds details about the current weather forecast
type Forecast struct {
	Current struct {
		ID            string    `json:"ID"`
		Name          string    `json:"Name"`
		Temp          float64   `json:"Temp"`
		Pressure      float64   `json:"Pressure"`
		Humidity      int       `json:"Humidity"`
		WindSpeed     float64   `json:"WindSpeed"`
		WindDirection float64   `json:"WindDirection"`
		Icon          string    `json:"Icon"`
		ReadingTime   time.Time `json:"ReadingTime"`
		Sunrise       time.Time `json:"Sunrise"`
		Sunset        time.Time `json:"Sunset"`
	} `json:"Current"`
	Days []struct {
		Day     time.Time `json:"Day"`
		Name    string    `json:"Name"`
		TempMin float64   `json:"TempMin"`
		TempMax float64   `json:"TempMax"`
		Icon    string    `json:"Icon"`
	} `json:"Days"`
}

// GetForecast returns the current weather forecast
func GetForecast() (Forecast, error) {
	f := Forecast{}
	resp, err := http.Get("http://localhost:20511/weather/forecast")
	if err == nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(b, &f)
		}
	}

	return f, err
}
