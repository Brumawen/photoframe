package main

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	gopifinder "github.com/brumawen/gopi-finder/src"
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

	// Wait until we have an internet connection
	a := 0
	for !gopifinder.IsInternetOnline() {
		a = a + 1
		if a == 5 {
			d.logInfo("Timeout waiting for internet.")
			return
		}
		d.logInfo("Waiting for internet connection.")
		time.Sleep(time.Minute)
	}

	switch d.Srv.Config.Resolution {
	case 0: // 800x480
		d.xBlock = 200
		d.yBlock = 120
	}

	// Get the list of images
	var l []DisplayImage
	p, n, err := d.getImageProvider()
	if err != nil {
		d.logError("Error getting image provider. ", err.Error())
		d.LastErr = err
		return
	}
	l, err = p.GetImages()
	if err != nil {
		d.logError("Error getting images from", n, ". ", err.Error())
		d.LastErr = err
		return
	}

	// Get the current weather forecast
	w := Weather{}
	m := Moon{}
	if d.Srv.Config.Weather {
		d.logInfo("Getting weather forecast.")
		w, err = GetForecast()
		if err != nil {
			d.logError("Error getting weather forecast.", err.Error())
			d.LastErr = err
			return
		}

		// Get the current moon phase
		d.logInfo("Getting moon information.")
		m, err = GetMoon()
		if err != nil {
			d.logError("Error getting moon phase details.", err.Error())
			d.LastErr = err
			return
		}
	}

	// Get the Calendar events
	c := CalEvents{}
	if d.Srv.Config.Calendar {
		d.logInfo("Getting calendar event information.")
		c, err = GetCalendarEvents()
		if err != nil {
			d.logInfo("Error getting calendar event information.", err.Error())
			d.LastErr = err
			return
		}
	}

	// Process the images
	d.logInfo("Building display images.")
	dl, err := d.buildDisplayImages(l, w, m, c)
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
				n := f.Name()
				if n != "System Volume Information" {
					p := filepath.Join(d.Srv.Config.USBPath, f.Name())
					err = os.Remove(p)
					if err != nil {
						d.logError(fmt.Sprintf("Error removing file '%s'", p))
					}
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

	// Refresh the USB mount
	d.RefreshUSB()

	d.IsRunning = false
	d.LastErr = nil
	d.logInfo("Processing complete.")
}

// RefreshUSB will remove and readd the USB mount, so that the display will trigger a reload of the images
func (d *Display) RefreshUSB() {
	go func() {
		d.logInfo("Refreshing USB display")
		myInfo, err := gopifinder.NewDeviceInfo()
		if err != nil {
			d.logError("Error getting device information. " + err.Error())
			return
		}
		if myInfo.OS != "Linux" {
			d.logInfo("Refresh of USB is not supported on " + myInfo.OS)
			return
		}
		// Switch off the USB
		d.logDebug("Removing USB entry.")
		err = exec.Command("sudo", "modprobe", "-r", "g_mass_storage").Run()
		if err != nil {
			d.logError("Error removing USB entry. " + err.Error())
		}
		d.logDebug("Adding USB entry.")
		err = exec.Command("sudo", "modprobe", "g_mass_storage", "file=/piusb.bin", "stall=0", "ro=1").Run()
		if err != nil {
			d.logError("Error adding USB entry. " + err.Error())
		}
		d.logInfo("Refresh USB display complete.")
	}()
}

func (d *Display) getImageProvider() (ImageProvider, string, error) {
	switch d.Srv.Config.Provider {
	case 0:
		n := "Bing Image of the Day"
		p := new(IodBing)
		p.SetConfig(*d.Srv.Config)
		return p, n, nil
	case 1:
		n := "Lorem Picsum"
		p := new(LoremPicsum)
		p.SetConfig(*d.Srv.Config)
		return p, n, nil
	default:
		n := "Unknown Image Provider"
		return nil, n, fmt.Errorf("Image Provider '%d' is invalid", d.Srv.Config.Provider)
	}
}

func (d *Display) buildDisplayImages(dl []DisplayImage, w Weather, m Moon, c CalEvents) ([]DisplayImage, error) {
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

	if d.Srv.Config.Weather || d.Srv.Config.Calendar {
		n := 0
		for _, i := range dl {
			if d.Srv.Config.Weather {
				if img, err := d.buildWeatherImage(n, i, w, m); err == nil {
					rl = append(rl, img)
					n = n + 1
				}
			}
			if d.Srv.Config.Calendar {
				if img, err := d.buildCalendarImage(n, i, c); err == nil {
					rl = append(rl, img)
					n = n + 1
				}
			}
		}
	} else {
		rl = dl
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

func (d *Display) buildCalendarImage(n int, i DisplayImage, c CalEvents) (DisplayImage, error) {
	// Load the image
	di := DisplayImage{Name: i.Name}
	img, err := gg.LoadImage(i.ImagePath)
	if err != nil {
		d.logError("Error loading image " + i.ImagePath + " - " + err.Error())
		return di, err
	}

	// Create a context for the image
	dc := gg.NewContextForImage(img)

	// Draw the day names
	now := time.Now()
	yer := now.Year()
	mth := now.Month()
	day := now.Day()
	now = time.Date(yer, mth, day, 0, 0, 0, 0, time.Local)
	cd := now

	for i := 0; i < 4; i++ {
		xb := i*d.xBlock + 10
		d.drawString(dc, cd.Weekday().String(), 20, xb, 10)
		cd = cd.Add(24 * time.Hour)
	}
	_, h := dc.MeasureString(now.Weekday().String())
	ht := int(h + 30)

	xq := 0
	y := ht
	// Draw the events onto the image
	cdn := ""
	for _, e := range c {
		if cdn == "" || cdn != e.DayName {
			cdn = e.DayName
			xq = int(e.Start.Sub(now).Hours() / 24)
			y = ht
		}
		y = d.drawCalEvent(dc, e, xq, y)
	}

	d.drawCalNames(dc, 3, 3)

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

func (d *Display) drawCalEvent(dc *gg.Context, e CalEvent, xq int, y int) int {
	xb := xq*d.xBlock + 15
	gap := 12
	fsize := 20

	if err := dc.LoadFontFace("./html/assets/font/Roboto-Black.ttf", float64(fsize)); err != nil {
		d.logError("Error loading font. " + err.Error())
	}

	t := fmt.Sprintf("%s (%s)", e.Time, e.Duration)

	w, h := dc.MeasureString(t)
	max := float64(d.xBlock - 10)
	for w > max {
		t = t[:len(t)-1]
		w, h = dc.MeasureString(t)
	}
	t = strings.TrimSpace(t)

	d.drawColourString(dc, t, fsize, e.Colour, int(xb), y)

	y = y + int(h) + 5

	t = fmt.Sprintf(" %s", e.Summary)

	w, h = dc.MeasureString(t)
	max = float64(d.xBlock - 10)
	for w > max {
		t = t[:len(t)-1]
		w, h = dc.MeasureString(t)
	}
	t = strings.TrimSpace(t)

	d.drawColourString(dc, t, fsize, e.Colour, int(xb), y)

	return y + int(h) + gap
}

func (d *Display) drawCalNames(dc *gg.Context, xq int, yq int) {
	nl, err := GetCalendarNames()
	if err != nil {
		d.logError("Failed to get Calendar Names. " + err.Error())
	} else {
		fsize := 20

		if err := dc.LoadFontFace("./html/assets/font/Roboto-Black.ttf", float64(fsize)); err != nil {
			d.logError("Error loading font. " + err.Error())
		}

		yb := d.yBlock * yq
		xb := d.xBlock * xq
		for _, name := range nl {
			_, h := dc.MeasureString(name.Name)
			d.drawColourString(dc, name.Name, fsize, name.Colour, xb, yb)
			yb = yb + int(h) + 15
		}
	}
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
	if int(dc.FontHeight()) != h {
		if err := dc.LoadFontFace("./html/assets/font/Roboto-Black.ttf", float64(h)); err != nil {
			d.logError("Error loading font. " + err.Error())
		}
	}
	_, sh := dc.MeasureString(s)

	dc.SetColor(color.Black)
	dc.DrawString(s, float64(x+1), float64(y+1)+sh)

	dc.SetColor(color.White)
	dc.DrawString(s, float64(x), float64(y)+sh)
}

func (d *Display) drawColourString(dc *gg.Context, s string, h int, c string, x int, y int) {
	if int(dc.FontHeight()) != h {
		if err := dc.LoadFontFace("./html/assets/font/Roboto-Black.ttf", float64(h)); err != nil {
			d.logError("Error loading font. " + err.Error())
		}
	}
	_, sh := dc.MeasureString(s)

	dc.SetColor(color.Black)
	dc.DrawString(s, float64(x+1), float64(y+1)+sh)

	dc.SetColor(d.getColour(c))
	dc.DrawString(s, float64(x), float64(y)+sh)
}

func (d *Display) getColour(c string) color.Color {
	switch c {
	case "Red":
		return color.RGBA{255, 0, 0, 255}
	case "Orange":
		return color.RGBA{255, 165, 0, 255}
	case "Yellow":
		return color.RGBA{255, 255, 0, 255}
	case "Tan":
		return color.RGBA{210, 180, 140, 255}
	case "Chocolate":
		return color.RGBA{210, 105, 30, 255}
	case "Lime":
		return color.RGBA{0, 255, 0, 255}
	case "SkyBlue":
		return color.RGBA{135, 206, 235, 255}
	case "Violet":
		return color.RGBA{238, 130, 238, 255}
	case "LightPink":
		return color.RGBA{255, 182, 193, 255}
	default:
		return color.White
	}
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
