package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Serge-Nook/sd-on-radio/internal/settings"
	"github.com/Serge-Nook/sd-on-radio/internal/ui"
)

var version = "2.0.0"

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("SD-ON RADIO %s\n", version)
		return
	}

	settingsPath, err := settings.Path()
	if err != nil {
		log.Fatal(err)
	}
	application, err := ui.New(settingsPath)
	if err != nil {
		log.Fatal(err)
	}
	if os.Getenv("SD_ON_RADIO_SMOKE_TEST") == "1" {
		time.AfterFunc(3*time.Second, application.Quit)
	}
	application.Run()
}
