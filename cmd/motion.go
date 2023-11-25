package main

import (
	"fmt"
	"log"
	"time"

	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
)

func main() {
	chipIdx := 0
	lineId := rpi.GPIO17
	line, err := gpiod.RequestLine(fmt.Sprintf("gpiochip%d", chipIdx), lineId)
	if err != nil {
		log.Fatal(err)
	}
	var prevVal int
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		<-ticker.C
		t := time.Now()
		//fmt.Printf("\r%s", t.Format("15:04:05.999"))
		val, err := line.Value()
		if err != nil {
			log.Fatal(err)
		}
		if val == 1 && prevVal == 0 {
			fmt.Println(t.Format("15:04:05.999"))
		}
		/*
		if val != prevVal {
			fmt.Printf("  %d\n", val)
		}
		*/
		prevVal = val
	}
}
