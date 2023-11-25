package types

import (
	"time"
)

type MotionSensorStatus struct {
	Now         time.Time   `json:"now"`
	LastMotion  time.Time   `json:"last_motion"`
	ElapsedTime float64     `json:"elapsed_time"`
	MotionLog   []time.Time `json:"motion_log"`
}
