package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FileFolder is an image provider that selects images from
// a configured folder on the disk
type FileFolder struct {
	Config Config
}

// SetConfig sets the configuration for this provider
func (p *FileFolder) SetConfig(c Config) {
	p.Config = c
}

// GetImages returns a slice of images to be used for display
func (p *FileFolder) GetImages() ([]DisplayImage, error) {
	l := []DisplayImage{}

	path := "./img/filefolder"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path does not exist, create it
		p.LogInfo(fmt.Sprintf("Creating path '%s'", path))
		err = os.MkdirAll(path, 0666)
		if err != nil {
			return l, err
		}
	}

	// Read the files in this folder
	fi, err := ioutil.ReadDir(path)
	if err == nil {
		for _, f := range fi {
			fp := filepath.Join(path, f.Name())
			l = append(l, DisplayImage{
				Name:      f.Name(),
				ImagePath: fp,
			})
		}
	}

	return l, err
}

// LogInfo is used to log information messages for this controller.
func (p *FileFolder) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v...)
	if logger != nil {
		logger.Info("NatGeo: [Inf] ", a)
	} else {
		fmt.Println("NatGeo: [Inf] ", a)
	}
}
