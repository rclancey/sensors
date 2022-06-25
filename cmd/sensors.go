package main

import (
	"math/rand"
	"time"

	"github.com/rclancey/sensors/api"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	api.APIMain()
}
