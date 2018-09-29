package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

type natgeoData struct {
	GalleryTitle     string `json:"galleryTitle"`
	PreviousEndpoint string `json:"previousEndpoint"`
	Items            []struct {
		Title       string  `json:"title"`
		Credit      string  `json:"credit"`
		ProfileURL  string  `json:"profileUrl"`
		AspectRatio float64 `json:"aspectRatio"`
		Sizes       struct {
			Num240  string `json:"240"`
			Num320  string `json:"320"`
			Num500  string `json:"500"`
			Num640  string `json:"640"`
			Num800  string `json:"800"`
			Num1024 string `json:"1024"`
			Num1600 string `json:"1600"`
			Num2048 string `json:"2048"`
		} `json:"sizes"`
	} `json:"items"`
}

// NatGeo is an image provider that selects images from
// National Georgraphic Image of the Day
type NatGeo struct {
	Config Config
}

// SetConfig sets the configuration for this provider
func (p *NatGeo) SetConfig(c Config) {
	p.Config = c
}

// GetImages returns a slice of images to be used for display
func (p *NatGeo) GetImages() ([]DisplayImage, error) {
	p.LogInfo("Downloading images from National Geographic.")

	l := []DisplayImage{}
	// Get the data from the National Geographic site
	url := "https://www.nationalgeographic.com/photography/photo-of-the-day/_jcr_content/.gallery.json"
	res, err := http.Get(url)
	if res != nil {
		defer res.Body.Close()
		res.Close = true
	}
	if err == nil {
		j, err := ioutil.ReadAll(res.Body)
		if err == nil {
			ngd := natgeoData{}
			err = json.Unmarshal(j, &ngd)
			if err == nil {
				l, err = p.downloadImages(&ngd)
			}
		}

	}

	if err != nil {
		// Check to see if we already have the last response cached
		fn := "lastnatgeo.json"
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

func (p *NatGeo) downloadImages(ngd *natgeoData) ([]DisplayImage, error) {
	l := []DisplayImage{}
	path := "./img/natgeo"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path does not exist, create it
		p.LogInfo(fmt.Sprintf("Creating path '%s'", path))
		err = os.MkdirAll(path, 0666)
		if err != nil {
			return l, err
		}
	}

	xRes, yRes := p.Config.GetResolution()

	for n, i := range ngd.Items {
		id := p.getImageID(i.ProfileURL)
		fn := fmt.Sprintf("%s.jpg", id)
		fp := filepath.Join(path, fn)
		load := true
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			url := ""
			switch p.Config.Resolution {
			case 0: // 800x480
				url = i.Sizes.Num800
			}
			p.LogInfo("Downloading", i.Title, fn)
			err = p.downloadImage(fp, fn, url, xRes, yRes)
			if err != nil {
				load = false
			}
		}
		if load {
			l = append(l, DisplayImage{
				Name:      fn,
				Copyright: fmt.Sprintf("%s - %s", i.Title, i.Credit),
				ImagePath: fp,
			})
		}
		if n+1 == p.Config.ImgCount {
			break
		}
	}

	return l, nil
}

func (p *NatGeo) downloadImage(fp string, fn string, url string, xRes int, yRes int) error {
	res, err := http.Get(url)
	if res != nil {
		defer res.Body.Close()
		res.Close = true
	}
	if err != nil {
		p.LogError("Error getting image file from url", url, ". ", err.Error())
		return err
	}
	fd, err := ioutil.ReadAll(res.Body)
	if err != nil {
		p.LogError("Error reading image file from response body.", err.Error())
		return err
	}
	err = ioutil.WriteFile(fp, fd, 0666)
	if err != nil {
		p.LogError("Error writing image file.", err.Error())
		return err
	}
	img, err := imaging.Open(fp)
	if err != nil {
		p.LogError("Error opening image for resizing.", err.Error())
	} else {
		img = imaging.Fill(img, xRes, yRes, imaging.Center, imaging.Lanczos)
		err = imaging.Save(img, fp)
	}

	return err
}

func (p *NatGeo) getImageID(url string) string {
	a := strings.Split(url, "/")
	l := len(a)
	if a[l-1] != "" {
		return a[l-1]
	}
	return a[l-2]
}

// LogInfo is used to log information messages for this controller.
func (p *NatGeo) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	if logger != nil {
		logger.Info("NatGeo: [Inf] ", a[1:len(a)-1])
	} else {
		fmt.Println("NatGeo: [Inf] ", a[1:len(a)-1])
	}
}

// LogError is used to log error messages for this controller.
func (p *NatGeo) LogError(v ...interface{}) {
	a := fmt.Sprint(v)
	if logger != nil {
		logger.Info("NatGeo: [Err] ", a[1:len(a)-1])
	} else {
		fmt.Println("NatGeo: [Err] ", a[1:len(a)-1])
	}
}
