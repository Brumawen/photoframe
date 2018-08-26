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

	res := ""
	switch b.Config.Resolution {
	case 0: // 800x480
		res = "_800x600"
	}

	for _, i := range bd.Images {
		// Check to see if the file already exists
		fn := filepath.Base(i.Urlbase) + ".jpg"
		fp := filepath.Join(path, fn)
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			// File does not exist, so download it
			b.LogInfo("Downloading", fn)
			res, err := http.Get("https://bing.com" + i.Urlbase + res + ".jpg")
			if err != nil {
				return l, err
			}
			fd, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return l, err
			}
			// Write the image data to the file
			err = ioutil.WriteFile(fp, fd, 0666)
			if err != nil {
				return l, err
			}
		}
		// Resize the image if required
		switch b.Config.Resolution {
		case 0: // 800x480
			// Resize the images from 800x600 to 800x480
			if img, err := imaging.Open(fp); err != nil {
				b.LogError("Error opening image for resizing. " + err.Error())
			} else {
				img = imaging.Fill(img, 800, 480, imaging.Center, imaging.Lanczos)
				imaging.Save(img, fp)
			}
		}
		l = append(l, DisplayImage{
			Name:      fn,
			ImagePath: fp,
		})
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
				b.LogInfo("Removing", f.Name())
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

// LogInfo is used to log information messages for this controller.
func (b *IodBing) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	if logger != nil {
		logger.Info("IodBing: [Inf] ", a[1:len(a)-1])
	} else {
		fmt.Println("IodBing: [Inf] ", a[1:len(a)-1])
	}
}

// LogError is used to log error messages for this controller.
func (b *IodBing) LogError(v ...interface{}) {
	a := fmt.Sprint(v)
	if logger != nil {
		logger.Info("IodBing: [Err] ", a[1:len(a)-1])
	} else {
		fmt.Println("IodBing: [Err] ", a[1:len(a)-1])
	}
}
