package main

import (
	"log"
	"time"

	"github.com/rclancey/kasa"
)


func main() {
	seen := map[string]bool{}
	for i := 0; i < 30; i++ {
		log.Println("attempt", i+1)
		devs, err := kasa.Discover(time.Duration(i+1) * time.Second)
		if err != nil {
			log.Println(err)
			return
		}
		for _, dev := range devs {
			if !seen[dev.IP()] {
				log.Println(dev.IP(), dev.MAC())
				seen[dev.IP()] = true
			}
		}
	}
}
