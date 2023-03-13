package main

import (
	"fmt"
	"testing"
)

func TestCanGetMoon(t *testing.T) {
	c := Config{}
	c.ReadFromFile("config.json")

	m, err := GetMoon(c)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(m)
}
