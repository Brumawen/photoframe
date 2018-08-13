package main

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

// Display is used to redraw the display images
type Display struct {
	Srv       *Server   // Server object
	LastRun   time.Time // Last run time
	IsRunning bool      // Indicates if the display build is running
	LastErr   error     // Last error encountered
	xBlock    int       // x block width
	yBlock    int       // y block height
}

// Run is called from the scheduler (ClockWerk).
func (d *Display) Run() {
	var err error

	d.logInfo("Starting Processing.")

	d.IsRunning = true

	switch d.Srv.Config.Resolution {
	case 0: // 800x480
		d.xBlock = 200
		d.yBlock = 120
	}

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
		d.LastErr = err
		return
	}

	// Get the current weather forecast
	d.logDebug("Getting weather forecast.")
	w, err := GetForecast()
	if err != nil {
		d.logError("Error getting weather forecast.", err.Error())
		d.LastErr = err
		return
	}

	// Get the current moon phase
	d.logDebug("Getting moon information.")
	m, err := GetMoon()
	if err != nil {
		d.logError("Error getting moon phase details.", err.Error())
		d.LastErr = err
		return
	}

	// Process the images
	d.logDebug("Building display images.")
	dl, err := d.buildDisplayImages(l, w, m)
	if err != nil {
		d.logError("Error building display images.", err.Error())
		d.LastErr = err
		return
	}

	// Check if the USB folder, where the files for display will be pulled from, exists
	if _, err := os.Stat(d.Srv.Config.USBPath); os.IsNotExist(err) {
		d.logError(fmt.Sprintf("Folder '%s' does not exist.", d.Srv.Config.USBPath))
	} else {
		d.logDebug("Moving images to the USB folder.")
		// Clear this folder
		if fi, err := ioutil.ReadDir(d.Srv.Config.USBPath); err == nil {
			for _, f := range fi {
				p := filepath.Join(d.Srv.Config.USBPath, f.Name())
				err = os.Remove(p)
				if err != nil {
					d.logError(fmt.Sprintf("Error removing file '%s'", p))
				}
			}
		}

		// Move the image files to the folder for display on the Photo Frame
		for _, i := range dl {
			n := filepath.Base(i.ImagePath)
			n = strings.TrimSuffix(n, path.Ext(n)) + ".jpg"
			p := filepath.Join(d.Srv.Config.USBPath, n)
			d.logDebug("Translating image " + p)
			if img, err := imaging.Open(i.ImagePath); err != nil {
				d.logError("Failed to open image for translation. " + err.Error())
			} else {
				imaging.Save(img, p)
			}
		}
	}

	d.IsRunning = false
	d.LastErr = nil
	d.logInfo("Processing complete.")
}

func (d *Display) buildDisplayImages(dl []DisplayImage, w Weather, m Moon) ([]DisplayImage, error) {
	rl := []DisplayImage{}

	// Clear the folder first
	path := "./img/display"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path does not exist, create it
		d.logInfo(fmt.Sprintf("Creating path '%s'", path))
		err = os.MkdirAll(path, 0666)
		if err != nil {
			d.logError(fmt.Sprintf("Creating path '%s'. %s", path, err.Error()))
			return rl, err
		}
	} else {
		if fi, err := ioutil.ReadDir(path); err == nil {
			for _, f := range fi {
				err = os.Remove(filepath.Join(path, f.Name()))
				if err != nil {
					d.logError(fmt.Sprintf("Error removing file '%s'", f.Name()))
				}
			}
		}
	}

	for n, i := range dl {
		if img, err := d.buildWeatherImage(n, i, w, m); err == nil {
			rl = append(rl, img)
		}
	}

	return rl, nil
}

func (d *Display) buildWeatherImage(n int, i DisplayImage, w Weather, m Moon) (DisplayImage, error) {
	// Load the image
	di := DisplayImage{Name: i.Name}
	img, err := gg.LoadImage(i.ImagePath)
	if err != nil {
		d.logError("Error loading image " + i.ImagePath + " - " + err.Error())
		return di, err
	}

	// Create a context for the image
	dc := gg.NewContextForImage(img)

	// Draw the sections
	d.drawCurrentTemp(dc, w, 0, 0)
	d.drawHumidPressure(dc, w, 0, 1)
	d.drawSunRiseSet(dc, w, 1, 1)
	d.drawWind(dc, w, 2, 1)
	d.drawMoon(dc, m, 2, 0)
	x := 0
	for i, f := range w.Forecast {
		if i <= 4 {
			if f.Day.YearDay() != time.Now().YearDay() {
				x = x + 1
				d.drawForecast(dc, w, i, 3, x-1)
			}
		}
	}

	// Save the new image
	di.ImagePath = filepath.Join("./img/display", fmt.Sprintf("image%d.png", n))
	err = dc.SavePNG(di.ImagePath)
	if err != nil {
		d.logError("Error saving weather image. " + err.Error())
	}

	return di, err
}

func (d *Display) drawCurrentTemp(dc *gg.Context, w Weather, xq int, yq int) {
	xb := xq*d.xBlock + 15
	yb := yq*d.yBlock + 10
	// Draw the icon
	if img, err := d.getWeatherIconImage(w.Current.WeatherIcon); err == nil {
		dc.DrawImage(img, xb, yb)
	}
	// Draw the weather description
	if w.Current.WeatherDesc != "" {
		d.drawString(dc, w.Current.WeatherDesc, 24, xb+10, yb+70)
	}
	// Draw the temperature
	temp := fmt.Sprintf("%.1f", w.Current.Temp)
	d.drawString(dc, temp, 50, xb+100, yb+10)
}

func (d *Display) drawHumidPressure(dc *gg.Context, w Weather, xq int, yq int) {
	xb := xq*d.xBlock + 15
	yb := yq * d.yBlock

	// Draw the Humidity icon
	if img, err := gg.LoadImage("./html/assets/images/humidity.png"); err == nil {
		dc.DrawImage(img, xb, yb)
	}
	// Draw the humidity value
	h := fmt.Sprintf("%.1f", w.Current.Humidity)
	d.drawString(dc, h, 20, xb+60, yb+12)

	yb = yb + 55

	// Draw the Pressure icon
	if img, err := gg.LoadImage("./html/assets/images/pressure.png"); err == nil {
		dc.DrawImage(img, xb+4, yb)
	}
	// Draw the pressure value
	p := fmt.Sprintf("%.1f", w.Current.Pressure)
	d.drawString(dc, p, 20, xb+60, yb+12)

}

func (d *Display) drawSunRiseSet(dc *gg.Context, w Weather, xq int, yq int) {
	xb := xq * d.xBlock
	yb := yq * d.yBlock

	// Draw the sunrise icon
	if img, err := gg.LoadImage("./html/assets/images/sunrise.png"); err == nil {
		dc.DrawImage(img, xb, yb)
	}
	// Draw the sunrise time
	t := w.Current.Sunrise.Format("3:04PM")
	d.drawString(dc, t, 20, xb+60, yb+12)

	yb = yb + 55

	// Draw the sunset icon
	if img, err := gg.LoadImage("./html/assets/images/sunset.png"); err == nil {
		dc.DrawImage(img, xb, yb)
	}
	// Draw the sunset time
	t = w.Current.Sunset.Format("3:04PM")
	d.drawString(dc, t, 20, xb+60, yb+12)
}

func (d *Display) drawWind(dc *gg.Context, w Weather, xq int, yq int) {
	xb := xq * d.xBlock
	yb := yq * d.yBlock

	// Draw the wind icon in the correct direction
	if img, err := gg.LoadImage("./html/assets/images/up.png"); err == nil {
		newImg := imaging.Rotate(img, float64(360-w.Current.WindDirection), color.Transparent)
		dc.DrawImage(newImg, xb, yb)
	}
	// Draw the wind speed value
	s := fmt.Sprintf("%.1f", w.Current.WindSpeed)
	d.drawString(dc, s, 20, xb+60, yb+12)
}

func (d *Display) drawMoon(dc *gg.Context, m Moon, xq int, yq int) {
	xb := xq * d.xBlock
	yb := yq * d.yBlock

	// Draw the moon icon
	if img, err := d.getMoonIconImage(m.Age); err == nil {
		dc.DrawImage(img, xb, yb+10)
	}
	// Draw the moon description
	if m.PhaseName != "" {
		d.drawString(dc, m.PhaseName, 15, xb+10, yb+70)
	}
}

func (d *Display) drawForecast(dc *gg.Context, w Weather, i int, xq int, yq int) {
	xb := xq*d.xBlock - 15
	yb := yq * d.yBlock
	fd := w.Forecast[i]

	// Draw the icon
	if img, err := d.getWeatherIconImage(fd.WeatherIcon); err == nil {
		dc.DrawImage(img, xb-10, yb)
	}
	// Draw the weather description
	if fd.WeatherDesc != "" {
		d.drawString(dc, fd.WeatherDesc, 15, xb+10, yb+70)
	}
	// Draw the day name
	if fd.Name != "" {
		d.drawString(dc, fd.Name, 18, xb+100, yb+10)
	}
	// Draw the temperature
	temp := fmt.Sprintf("%.0f / %.0f", fd.TempMax, fd.TempMin)
	d.drawString(dc, temp, 18, xb+100, yb+40)
}

func (d *Display) getWeatherIconImage(i int) (image.Image, error) {
	fn := ""
	switch i {
	case 1:
		fn = "sun1.png"
	case 2:
		fn = "suncloud1.png"
	case 3:
		fn = "cloud1.png"
	case 4:
		fn = "cloudy1.png"
	case 5:
		fn = "sunrain1.png"
	case 6:
		fn = "rain1.png"
	case 7:
		fn = "thunder1.png"
	case 8:
		fn = "snow1.png"
	case 9:
		fn = "mist1.png"
	default:
		fn = "unkown1.png"
	}

	p := filepath.Join("./html/assets/images", fn)
	return gg.LoadImage(p)
}

func (d *Display) getMoonIconImage(i float32) (image.Image, error) {
	fn := fmt.Sprintf("moon50_%d.png", int(i))
	p := filepath.Join("./html/assets/images", fn)
	return gg.LoadImage(p)
}

func (d *Display) drawString(dc *gg.Context, s string, h int, x int, y int) {
	if err := dc.LoadFontFace("./html/assets/font/Roboto-Black.ttf", float64(h)); err != nil {
		d.logError("Error loading font. " + err.Error())
	}
	_, sh := dc.MeasureString(s)

	dc.SetColor(color.Black)
	dc.DrawString(s, float64(x+1), float64(y+1)+sh)

	dc.SetColor(color.White)
	dc.DrawString(s, float64(x), float64(y)+sh)
}

func (d *Display) logDebug(v ...interface{}) {
	if d.Srv.VerboseLogging {
		a := fmt.Sprint(v)
		if logger != nil {
			logger.Info("Display: [Dbg] ", a[1:len(a)-1])
		} else {
			fmt.Println("Display: [Dbg] ", a[1:len(a)-1])
		}
	}
}

func (d *Display) logInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	if logger != nil {
		logger.Info("Display: [Inf] ", a[1:len(a)-1])
	} else {
		fmt.Println("Display: [Inf] ", a[1:len(a)-1])
	}
}

func (d *Display) logError(v ...interface{}) {
	a := fmt.Sprint(v)
	if logger != nil {
		logger.Error("Display: [Err] ", a[1:len(a)-1])
	} else {
		fmt.Println("Display: [Err] ", a[1:len(a)-1])
	}
}
