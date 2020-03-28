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
		Image struct {
			Title       string  `json:"title"`
			Credit      string  `json:"credit"`
			URI         string  `json:"uri"`
			AspectRatio float64 `json:"aspectRatio"`
		} `json:"image"`
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
	ngd, err := p.getNGData("https://www.nationalgeographic.com/photography/photo-of-the-day/_jcr_content/.gallery.json")

	if err == nil {
		// Check if we have enough data
		if len(ngd.Items) < p.Config.ImgCount && ngd.PreviousEndpoint != "" {
			ngd2, err := p.getNGData("https://www.nationalgeographic.com" + ngd.PreviousEndpoint)
			if err == nil {
				for _, i := range ngd2.Items {
					ngd.Items = append(ngd.Items, i)
				}
			}
		}
		// Download the images
		l, err = p.downloadImages(&ngd)
	}

	if err != nil {
		p.LogError("Error getting Images. ", err.Error())
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

func (p *NatGeo) getNGData(url string) (natgeoData, error) {
	// Get the data from the National Geographic site
	res, err := http.Get(url)
	if res != nil {
		defer res.Body.Close()
		res.Close = true
	}
	if err != nil {
		return natgeoData{}, err
	}
	j, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return natgeoData{}, err
	}
	ngd := natgeoData{}
	err = json.Unmarshal(j, &ngd)
	if err != nil {
		return natgeoData{}, err
	}
	return ngd, nil
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
		fn := p.getImageID(i.Image.URI)
		fp := filepath.Join(path, fn)
		load := true
		p.LogInfo("Checking image ", fp)
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			url := i.Image.URI
			if url == "" {
				load = false
			} else {
				p.LogInfo("Downloading", i.Image.Title, fn)
				err = p.downloadImage(fp, fn, url, xRes, yRes)
				if err != nil {
					load = false
				}
			}
		}
		if load {
			// Add the image to the list to return
			l = append(l, DisplayImage{
				Name:      fn,
				Copyright: fmt.Sprintf("%s - %s", i.Image.Title, i.Image.Credit),
				ImagePath: fp,
			})
		} else {
			// There was an issue processing the image,
			// remove the file from the disk if anything was written
			if _, err := os.Stat(fp); err == nil {
				p.LogInfo("Removing file", fp)
				os.Remove(fp)
			}
		}
		if n+1 == p.Config.ImgCount {
			break
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
					p.LogInfo("Error removing image file", f.Name(), ". ", err.Error())
				}
			}
		}
	}
	if err == nil {
		b, err := json.Marshal(l)
		if err != nil {
			return nil, err
		}
		ioutil.WriteFile("lastnatgeo.json", b, 0666)
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
		p.LogError("Error reading image file from response body.", url, ". ", err.Error())
		return err
	}
	err = ioutil.WriteFile(fp, fd, 0666)
	if err != nil {
		p.LogError("Error writing image file", fp, ". ", err.Error())
		return err
	}
	// Resize the image
	img, err := imaging.Open(fp)
	if err != nil {
		p.LogError("Error opening image file", fp, "for resizing.", err.Error())
	} else {
		img = imaging.Fill(img, xRes, yRes, imaging.Center, imaging.Lanczos)
		err = imaging.Save(img, fp)
		if err != nil {
			p.LogError("Error saving resized image file", fp, ".", err.Error())
		}
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
	a := fmt.Sprint(v...)
	if logger != nil {
		logger.Info("NatGeo: [Inf] ", a)
	} else {
		fmt.Println("NatGeo: [Inf] ", a)
	}
}

// LogError is used to log error messages for this controller.
func (p *NatGeo) LogError(v ...interface{}) {
	a := fmt.Sprint(v...)
	if logger != nil {
		logger.Info("NatGeo: [Err] ", a)
	} else {
		fmt.Println("NatGeo: [Err] ", a)
	}
}
