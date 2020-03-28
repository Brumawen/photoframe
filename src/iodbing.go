package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
)

type bingdata struct {
	Images []struct {
		URL       string `json:"url"`
		Urlbase   string `json:"urlbase"`
		Copyright string `json:"copyright"`
	}
}

// IodBing is used to retrieve the Bing images of the day
type IodBing struct {
	Config Config
}

// SetConfig sets the configuration for this provider
func (b *IodBing) SetConfig(c Config) {
	b.Config = c
}

// GetImages returns a slice of images to be used for display
func (b *IodBing) GetImages() ([]DisplayImage, error) {
	b.LogInfo("Getting latest list of images from Bing.")

	l := []DisplayImage{}
	// Get the data from the Bing web site
	resp, err := http.Get(fmt.Sprintf("http://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=%d&mkt=za", b.Config.ImgCount))
	if resp != nil {
		defer resp.Body.Close()
		resp.Close = true
	}
	if err == nil {
		j, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			bd := bingdata{}
			err = json.Unmarshal(j, &bd)
			if err == nil {
				l, err = b.downloadImages(&bd)
			}
		}
	}
	if err != nil {
		// Check to see if we already have the last response cached
		fn := "lastiodbing.json"
		if _, err := os.Stat(fn); !os.IsNotExist(err) {
			// Deserialize the last cached list
			b, err := ioutil.ReadFile(fn)
			if err == nil {
				err = json.Unmarshal(b, &l)
			}
		}
	}

	return l, err
}

func (b *IodBing) downloadImages(bd *bingdata) ([]DisplayImage, error) {
	b.LogInfo("Downloading images from Bing.")

	l := []DisplayImage{}
	path := "./img/bing"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path does not exist, create it
		b.LogInfo(fmt.Sprintf("Creating path '%s'", path))
		err = os.MkdirAll(path, 0666)
		if err != nil {
			return l, err
		}
	}

	xRes, yRes := b.Config.GetResolution()
	res := ""
	switch b.Config.Resolution {
	case 0: // 800x480
		res = "_800x600"
	}

	for _, i := range bd.Images {
		// Check to see if the file already exists
		fs := string([]rune(i.Urlbase)[7:])
		fn := filepath.Base(fs) + ".jpg"
		fp := filepath.Join(path, fn)
		load := true
		b.LogInfo("Checking if file '", fp, "' exits (", i.Urlbase, ")")
		_, err := os.Stat(fp)
		if os.IsNotExist(err) {
			// File does not exist, so download it
			b.LogInfo("Downloading", fp)
			url := "https://bing.com" + i.Urlbase + res + ".jpg"
			err = b.downloadImage(fp, fn, url, xRes, yRes)
			if err != nil {
				b.LogError("Failed with ", err.Error())
				load = false
			}

		} else if err != nil {
			b.LogError("Failed with ", err.Error())
			load = false
		}
		if load {
			// Add the image to the list to return
			l = append(l, DisplayImage{
				Name:      fn,
				Copyright: i.Copyright,
				ImagePath: fp,
			})
		} else {
			// There was an issue processing the image,
			// remove the file from the disk if anything was written
			if _, err := os.Stat(fp); err == nil {
				b.LogInfo("Removing file", fp)
				os.Remove(fp)
			}
		}
	}

	// Remove any other file in this folder
	fi, err := ioutil.ReadDir(path)
	if err == nil {
		for _, f := range fi {
			// Check if this file is in the list
			remove := true
			for _, i := range l {
				if i.Name == f.Name() {
					remove = false
					break
				}
			}
			if remove {
				b.LogInfo("Removing file '", f.Name(), "'")
				err = os.Remove(filepath.Join(path, f.Name()))
				if err != nil {
					b.LogInfo("Error", err.Error())
				}
			}
		}
	}
	if err == nil {
		b, err := json.Marshal(l)
		if err != nil {
			return nil, err
		}
		ioutil.WriteFile("lastiodbing.json", b, 0666)
	}

	return l, err
}

func (b *IodBing) downloadImage(fp string, fn string, url string, xRes int, yRes int) error {
	res, err := http.Get(url)
	if res != nil {
		defer res.Body.Close()
		res.Close = true
	}
	if err != nil {
		b.LogError("Error getting image file from url", url, ". ", err.Error())
		return err
	}
	fd, err := ioutil.ReadAll(res.Body)
	if err != nil {
		b.LogError("Error reading image file from response body.", url, ". ", err.Error())
		return err
	}
	b.LogInfo("Downloading file to ", fp)
	err = ioutil.WriteFile(fp, fd, 0666)
	if err != nil {
		b.LogError("Error writing image file", fp, ". ", err.Error())
		return err
	}
	// Resize the image
	img, err := imaging.Open(fp)
	if err != nil {
		b.LogError("Error opening image for resizing. " + err.Error())
	} else {
		img = imaging.Fill(img, xRes, yRes, imaging.Center, imaging.Lanczos)
		err = imaging.Save(img, fp)
		if err != nil {
			b.LogError("Error saving resized image file", fp, ".", err.Error())
		}
	}
	return err
}

// LogInfo is used to log information messages for this controller.
func (b *IodBing) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v...)
	if logger != nil {
		logger.Info("IodBing: [Inf] ", a)
	} else {
		fmt.Println("IodBing: [Inf] ", a)
	}
}

// LogError is used to log error messages for this controller.
func (b *IodBing) LogError(v ...interface{}) {
	a := fmt.Sprint(v...)
	if logger != nil {
		logger.Info("IodBing: [Err] ", a)
	} else {
		fmt.Println("IodBing: [Err] ", a)
	}
}
