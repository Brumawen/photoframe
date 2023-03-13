package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// Config holds the configuration required for the Soil Monitor module.
type Config struct {
	Resolution  int    `json:"resolution"`  // Resolution of the display, 0=800x480
	Provider    int    `json:"provider"`    // Image of the Day provider
	ImgCount    int    `json:"imgcount"`    // NUmber of images to retrieve
	Weather     bool   `json:"weather"`     // Display weather data
	WeatherUrl  string `json:"weatherurl"`  // Url for the weather service
	Calendar    bool   `json:"calendar"`    // Display calendar data
	Loadshed    bool   `json:"loadshed"`    // Display Load shedding data
	LoadshedUrl string `json:"loadshedurl"` // Url for the load shedding service
	USBPath     string `json:"usbPath"`     // Path to the USB shared folder
}

// GetResolution returns the required image resolution (x,y)
func (c *Config) GetResolution() (int, int) {
	xRes := 0
	yRes := 0

	switch c.Resolution {
	case 0: // 800x480
		xRes = 800
		yRes = 480
	}

	return xRes, yRes
}

// ReadFromFile will read the configuration settings from the specified file
func (c *Config) ReadFromFile(path string) error {
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(path)
		if err == nil {
			err = json.Unmarshal(b, &c)
		}
	}
	c.SetDefaults()
	return err
}

// WriteToFile will write the configuration settings to the specified file
func (c *Config) WriteToFile(path string) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0666)
}

// ReadFrom reads the string from the reader and deserializes it into the entity values
func (c *Config) ReadFrom(r io.ReadCloser) error {
	b, err := ioutil.ReadAll(r)
	if err == nil {
		if b != nil && len(b) != 0 {
			err = json.Unmarshal(b, &c)
		}
	}
	c.SetDefaults()
	return err
}

// WriteTo serializes the entity and writes it to the http response
func (c *Config) WriteTo(w http.ResponseWriter) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	w.Header().Set("content-type", "application/json")
	w.Write(b)
	return nil
}

// Serialize serializes the entity and returns the serialized string
func (c *Config) Serialize() (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Deserialize deserializes the specified string into the entity values
func (c *Config) Deserialize(v string) error {
	err := json.Unmarshal([]byte(v), &c)
	c.SetDefaults()
	return err
}

// SetDefaults checks the values and sets the defaults
func (c *Config) SetDefaults() {
	mustSave := false
	if c.ImgCount < 1 {
		c.ImgCount = 8
		mustSave = true
	}
	if c.USBPath == "" {
		c.USBPath = "/mnt/usb_share"
		mustSave = true
	}
	if mustSave {
		c.WriteToFile("config.json")
	}
	if c.WeatherUrl == "" {
		c.WeatherUrl = "http://localhost:20511"
		mustSave = true
	}
	if c.LoadshedUrl == "" {
		c.LoadshedUrl = "http://localhost:20515"
		mustSave = true
	}
}
