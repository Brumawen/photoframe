package main

// ImageProvider defines an interface for an Image provider
type ImageProvider interface {
	GetImages() ([]DisplayImage, error)
	SetConfig(c Config)
}
