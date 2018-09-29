package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type pexelsData struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Photos  []struct {
		ID           int    `json:"id"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		URL          string `json:"url"`
		Photographer string `json:"photographer"`
	} `json:"photos"`
}

// Pexels is an image provider that selects images from Pexels.com
type Pexels struct {
	Config Config
}

// SetConfig sets the configuration for this provider
func (p *Pexels) SetConfig(c Config) {
	p.Config = c
}

// GetImages returns a slice of images to be used for display
func (p *Pexels) GetImages() ([]DisplayImage, error) {
	p.LogInfo("Downloading images from Pexels.")

	l := []DisplayImage{}
	// Get the data from the Pexels API
	url := fmt.Sprintf("https://api.pexels.com/v1/curated?per_page=%d&page=1", p.Config.ImgCount)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "563492ad6f917000010000012b1ae72ef5bd4abb9ba8157e7b653f43")
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
		resp.Close = true
	}
	if err == nil {
		j, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			pd := pexelsData{}
			err = json.Unmarshal(j, &pd)
			if err == nil {
				l, err = p.downloadImages(&pd)
			}
		}
	}

	if err != nil {
		// Check to see if we already have the last response cached
		fn := "lastpexels.json"
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

func (p *Pexels) downloadImages(pd *pexelsData) ([]DisplayImage, error) {
	l := []DisplayImage{}
	path := "./img/pexels"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path does not exist, create it
		p.LogInfo(fmt.Sprintf("Creating path '%s'", path))
		err = os.MkdirAll(path, 0666)
		if err != nil {
			return l, err
		}
	}

	xRes, yRes := p.Config.GetResolution()

	for _, i := range pd.Photos {
		// Check if the file already exists
		fn := fmt.Sprintf("%d.jpg", i.ID)
		fp := filepath.Join(path, fn)
		load := true
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			err = p.downloadImage(fp, fn, i.ID, xRes, yRes)
			if err != nil {
				load = false
			}
		}
		if load {
			l = append(l, DisplayImage{
				Name:      fn,
				Copyright: i.Photographer,
				ImagePath: fp,
			})
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
				p.LogInfo("Removing", f.Name())
				err = os.Remove(filepath.Join(path, f.Name()))
				if err != nil {
					p.LogInfo("Error", err.Error())
				}
			}
		}
	}
	if err == nil {
		b, err := json.Marshal(l)
		if err != nil {
			return nil, err
		}
		ioutil.WriteFile("lastpexels.json", b, 0666)
	}

	return l, nil
}

func (p *Pexels) downloadImage(fp string, fn string, id int, xRes int, yRes int) error {
	// File does not exist, so download it
	p.LogInfo("Downloading", fn)
	url := fmt.Sprintf("https://images.pexels.com/photos/%d/pexels-photo-%d.jpeg?auto=compress&cs=tinysrgb&fit=crop&h=%d&w=%d", id, id, yRes, xRes)
	res, err := http.Get(url)
	if res != nil {
		defer res.Body.Close()
		res.Close = true
	}
	if err != nil {
		return err
	}
	fd, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fp, fd, 0666)
	return err
}

// LogInfo is used to log information messages for this controller.
func (p *Pexels) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	if logger != nil {
		logger.Info("Pexels: [Inf] ", a[1:len(a)-1])
	} else {
		fmt.Println("Pexels: [Inf] ", a[1:len(a)-1])
	}
}

// LogError is used to log error messages for this controller.
func (p *Pexels) LogError(v ...interface{}) {
	a := fmt.Sprint(v)
	if logger != nil {
		logger.Info("Pexels: [Err] ", a[1:len(a)-1])
	} else {
		fmt.Println("Pexels: [Err] ", a[1:len(a)-1])
	}
}
