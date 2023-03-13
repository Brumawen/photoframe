package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Moon holds the details about a moon phase
type Loadshed struct {
	Name   string `json:"name"`   // Area Name
	Region string `json:"region"` // Region Name
	Stage  int    `json:"stage"`  // Current Stage
	Events []struct {
		Start   time.Time `json:"start"` // Start date and time
		End     time.Time `json:"end"`   // End date and time
		Day     string    `json:"day"`   // The day of the event (e.g. Mon, Tue etc)
		Display string    `json:"note"`  // Display information
		Stage   int       `json:"stage"` // Stage
	}
}

// GetLoadshedInfo returns the current load shedding forecasts
func GetLoadshedInfo(c Config) (Loadshed, error) {
	m := Loadshed{}
	resp, err := http.Get(fmt.Sprintf("%s/forecast/get", c.LoadshedUrl))
	if resp != nil {
		defer resp.Body.Close()
		resp.Close = true
	}
	if err == nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(b, &m)
		}
	}

	if err == nil {
		m.WriteToFile("lastloadshed.json")
	}

	return m, err
}

// WriteToFile will write the forecast information to the specified file
func (f *Loadshed) WriteToFile(path string) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0666)
}
