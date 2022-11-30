package main

import (
	"encoding/json"
	"fmt"
	"time"

	gnss "github.com/Pingoin/gpsd-client"
)

func main() {
	gps := gnss.NewGPSD(gnss.DefaultAddress, 0, 0)
	gps.SetDebug(true)

	// wait for reply
	go gps.Start()

	for {
		time.Sleep(time.Second * 5)
		date, _ := json.MarshalIndent(gps, "", " ")
		fmt.Println(string(date))
	}
}
