package types

import (
	"time"

	"github.com/rclancey/kasa"

)

type Device struct {
	kasa.SmartDevice
	State      int
	LastUpdate time.Time
}
