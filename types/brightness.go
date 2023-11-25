package types

import (
	"time"

	"github.com/rclancey/sensors/tsl2591"
)

type BrightnessReading struct {
	*tsl2591.SensorData
	Now time.Time `json:"now"`
	Calculated float64 `json:"calculated"`
}

func (br *BrightnessReading) Value() float64 {
	return float64(br.Lux)
}
