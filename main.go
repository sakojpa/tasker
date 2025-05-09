package main

import (
	"github.com/sakojpa/tasker/cmd"
	"log"
)

func main() {
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
