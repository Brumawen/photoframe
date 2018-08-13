package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

// Weather holds the current weather forecast received from the Weather Micro-service
type Weather struct {
	Current struct {
		Provider      string    `json:"provider"`
		Created       time.Time `json:"created"`
		LocationID    string    `json:"locationID"`
		LocationName  string    `json:"locationName"`
		Temp          float32   `json:"temp"`
		Pressure      float32   `json:"pressure"`
		Humidity      float32   `json:"humidity"`
		WindSpeed     float32   `json:"windSpeed"`
		WindDirection float32   `json:"windDirection"`
		WeatherIcon   int       `json:"weatherIcon"`
		WeatherDesc   string    `json:"weatherDesc"`
		IsDay         bool      `json:"isDay"`
		ReadingTime   time.Time `json:"readingTime"`
		Sunrise       time.Time `json:"sunrise"`
		Sunset        time.Time `json:"sunset"`
	} `json:"current"`
	Forecast []struct {
		Day         time.Time `json:"day"`
		Name        string    `json:"name"`
		TempMin     float32   `json:"tempMin"`
		TempMax     float32   `json:"tempMax"`
		WeatherIcon int       `json:"weatherIcon"`
		WeatherDesc string    `json:"weatherDesc"`
	} `json:"forecast"`
}

// GetForecast returns the current weather forecast
func GetForecast() (Weather, error) {
	f := Weather{}
	resp, err := http.Get("http://localhost:20511/weather/forecast")
	if err == nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(b, &f)
		}
	}

	if err == nil {
		f.WriteToFile("lastweather.json")
	}

	return f, err
}

// WriteToFile will write the forecast information to the specified file
func (w *Weather) WriteToFile(path string) error {
	b, err := json.Marshal(w)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0666)
}
