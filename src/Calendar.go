package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

// CalEvents holds a list of calendar events
type CalEvents []CalEvent

// CalEvent holds details about a calendar event
type CalEvent struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	DayName     string    `json:"dayName"`
	Time        string    `json:"time"`
	Duration    string    `json:"duration"`
	Summary     string    `json:"summary"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	Colour      string    `json:"colour"`
}

// CalNames holds the names and colours of the configured calendars
type CalNames []struct {
	Name   string `json:"name"`
	Colour string `json:"colour"`
}

// GetCalendarNames returns the names and colours of the configured calendars
func GetCalendarNames() (CalNames, error) {
	c := CalNames{}
	resp, err := http.Get("http://localhost:20513/calendar/get")
	if err == nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(b, &c)
		}
	}

	if err == nil {
		c.WriteToFile("lastcalnames.json")
	}

	return c, err
}

// GetCalendarEvents returns the calendar events for the next 4 days
func GetCalendarEvents() (CalEvents, error) {
	c := CalEvents{}
	resp, err := http.Get("http://localhost:20513/calendar/get/4")
	if err == nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(b, &c)
		}
	}

	if err == nil {
		c.WriteToFile("lastcalevents.json")
	}

	return c, err
}

// WriteToFile will write the calendar name information to the specified file
func (c *CalNames) WriteToFile(path string) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0666)
}

// WriteToFile will write the calendar event information to the specified file
func (c *CalEvents) WriteToFile(path string) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0666)
}
