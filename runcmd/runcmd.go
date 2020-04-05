package main

import (
	"fmt"
	"os/exec"
)

func main() {
	err := exec.Command("./stopusb").Run()
	if err != nil {
		fmt.Println(err.Error())
	}
}
