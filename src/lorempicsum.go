package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

// LoremPicsum is an image provider that selects images from https://picsum.photos/
type LoremPicsum struct {
	Config Config
}

// SetConfig sets the configuration for this provider
func (p *LoremPicsum) SetConfig(c Config) {
	p.Config = c
}

// GetImages returns a slice of images to be used for display
func (p *LoremPicsum) GetImages() ([]DisplayImage, error) {
	p.LogInfo("Getting latest list of images from Lorem Picsum.")

	r := ""
	switch p.Config.Resolution {
	case 0: // 800x480
		r = "/800/480/"
	}

	l := []DisplayImage{}
	path := "./img/lorem"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path does not exist, create it
		p.LogInfo(fmt.Sprintf("Creating path '%s'", path))
		err = os.MkdirAll(path, 0666)
		if err != nil {
			return l, err
		}
	}

	for i := 0; i < p.Config.ImgCount; i++ {
		fn := fmt.Sprintf("image%d.jpg", i)
		fp := filepath.Join(path, fn)
		url := fmt.Sprintf("https://picsum.photos%s?random", r)
		res, err := http.Get(url)
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
		l = append(l, DisplayImage{
			Name:      fn,
			ImagePath: fp,
		})
	}
	return l, nil
}

// LogInfo is used to log information messages for this controller.
func (p *LoremPicsum) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	if logger != nil {
		logger.Info("LoremPicsum: [Inf] ", a[1:len(a)-1])
	} else {
		fmt.Println("LoremPicsum: [Inf] ", a[1:len(a)-1])
	}
}

// LogError is used to log error messages for this controller.
func (p *LoremPicsum) LogError(v ...interface{}) {
	a := fmt.Sprint(v)
	if logger != nil {
		logger.Info("LoremPicsum: [Err] ", a[1:len(a)-1])
	} else {
		fmt.Println("LoremPicsum: [Err] ", a[1:len(a)-1])
	}
}