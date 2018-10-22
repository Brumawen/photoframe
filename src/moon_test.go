package main

import (
	"fmt"
	"testing"
)

func TestCanGetMoon(t *testing.T) {
	m, err := GetMoon()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(m)
}
