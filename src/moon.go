package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

// Moon holds the details about a moon phase
type Moon struct {
	Date         time.Time `json:"Date"`
	Age          float64   `json:"Age"`
	Phase        float64   `json:"Phase"`
	PhaseName    string    `json:"PhaseName"`
	Illumination float64   `json:"Illumination"`
}

// GetMoon returns the details about the current phase of the moon
func GetMoon() (Moon, error) {
	m := Moon{}
	resp, err := http.Get("http://localhost:20511/moon/get")
	if err == nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(b, &m)
		}
	}

	return m, err
}
