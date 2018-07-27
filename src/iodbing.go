package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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

	res := "_800x600"

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
			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return l, err
			}
			// Write the image data to the file
			err = ioutil.WriteFile(fp, b, 0666)
			if err != nil {
				return l, err
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

	return l, nil
}

// LogInfo is used to log information messages for this controller.
func (b *IodBing) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	if logger != nil {
		logger.Info("IodBing: ", a[1:len(a)-1])
	} else {
		fmt.Println("IodBing: ", a[1:len(a)-1])
	}
}
