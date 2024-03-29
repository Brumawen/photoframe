package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Moon holds the details about a moon phase
type Moon struct {
	Date         time.Time `json:"Date"`
	Age          float32   `json:"Age"`
	Phase        float32   `json:"Phase"`
	PhaseName    string    `json:"PhaseName"`
	Illumination float32   `json:"Illumination"`
}

// GetMoon returns the details about the current phase of the moon
func GetMoon(c Config) (Moon, error) {
	m := Moon{}
	resp, err := http.Get(fmt.Sprintf("%s/moon/get", c.WeatherUrl))
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
		m.WriteToFile("lastmoon.json")
	}

	return m, err
}

// WriteToFile will write the forecast information to the specified file
func (m *Moon) WriteToFile(path string) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0666)
}
