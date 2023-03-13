package main

import (
	"fmt"
	"testing"
)

func TestCanGetLoadshedInfo(t *testing.T) {
	c := Config{}
	c.ReadFromFile("config.json")

	m, err := GetLoadshedInfo(c)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(m)
}
