package main

import (
	"fmt"
	"time"
)

// Display is used to redraw the display images
type Display struct {
	Srv       *Server
	LastRun   time.Time
	IsRunning bool
}

// Run is called from the scheduler (ClockWerk).
func (d *Display) Run() {
	var err error

	// Get the list of images
	var l []DisplayImage
	var p string
	if d.Srv.Config.Provider == 0 {
		p = "Bing Image of the Day"
		i := IodBing{Config: *d.Srv.Config}
		l, err = i.GetImages()
	}
	if err != nil {
		d.logError("Error getting images from", p, ". ", err.Error())
		return
	}

	// Get the current weather forecase
	w, err := GetForecast()
	if err != nil {
		d.logError("Error getting weather forecast.", err.Error())
		return
	}

	// Get the current moon phase
	m, err := GetMoon()
	if err != nil {
		d.logError("Error getting moon phase details.", err.Error())
		return
	}

	// Process the images
	_, err = d.buildDisplayImages(l, w, m)
	if err != nil {
		d.logError("Error building display images.", err.Error())
		return
	}

	// Move the image files to the folder for display on the Photo Frame

}

func (d *Display) buildDisplayImages(l []DisplayImage, w Forecast, m Moon) ([]DisplayImage, error) {
	di := []DisplayImage{}
	return di, nil
}

func (d *Display) logDebug(v ...interface{}) {
	if d.Srv.VerboseLogging {
		a := fmt.Sprint(v)
		logger.Info("Display: ", a[1:len(a)-1])
	}
}

func (d *Display) logInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Info("Display: ", a[1:len(a)-1])
}

func (d *Display) logError(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Error("Display: ", a[1:len(a)-1])
}
