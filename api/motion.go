package api

import (
	//"encoding/json"
	"fmt"
	"net/http"
	//"os/exec"
	//"path/filepath"
	"time"

	"github.com/rclancey/events"
	//"github.com/rclancey/generic"
	//"github.com/rclancey/httpserver/v2"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
)

type MotionSensorStatus struct {
	Now time.Time `json:"now"`
	LastMotion time.Time `json:"last_motion"`
	ElapsedTime float64 `json:"elapsed_time"`
	MotionLog []time.Time `json:"motion_log"`
}

type MotionSensor struct {
	cfg *Config
	eventSink events.EventSink
	line *gpiod.Line
	lastMotion time.Time
	lastStillness time.Duration
}

func NewMotionSensor(cfg *Config, eventSink events.EventSink) (*MotionSensor, error) {
	chipIdx := 0
	lineId := rpi.GPIO17
	line, err := gpiod.RequestLine(fmt.Sprintf("gpiochip%d", chipIdx), lineId)
	if err != nil {
		return nil, err
	}
	ms := &MotionSensor{
		cfg: cfg,
		eventSink: events.NewPrefixedEventSource("motion", eventSink),
		line: line,
		lastMotion: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.Local),
	}
	ms.registerEventTypes()
	return ms, nil
}

func (ms *MotionSensor) registerEventTypes() {
	now := time.Now()
	ms.eventSink.RegisterEventType(events.NewEvent("movement", &MotionSensorStatus{
		Now: now,
		LastMotion: now,
		ElapsedTime: 0,
	}))
	ms.eventSink.RegisterEventType(events.NewEvent("stillness", &MotionSensorStatus{
		Now: now,
		LastMotion: now.Add(-5 * time.Minute),
		ElapsedTime: 300,
	}))
}

func (ms *MotionSensor) Check() (interface{}, error) {
	val, err := ms.line.Value()
	if err != nil {
		return nil, err
	}
	status := &MotionSensorStatus{
		Now: time.Now().In(time.UTC),
		LastMotion: ms.lastMotion,
		ElapsedTime: 0,
	}
	if val != 0 {
		ms.lastMotion = status.Now
		ms.lastStillness = 0
		status.LastMotion = status.Now
	} else {
		elapsed := status.Now.Sub(status.LastMotion)
		status.ElapsedTime = elapsed.Seconds()
		durs := []time.Duration{
			time.Minute,
			5 * time.Minute,
			15 * time.Minute,
			30 * time.Minute,
			time.Hour,
			2 * time.Hour,
			4 * time.Hour,
			8 * time.Hour,
			12 * time.Hour,
			24 * time.Hour,
			36 * time.Hour,
			48 * time.Hour,
			60 * time.Hour,
			72 * time.Hour,
		}
		for _, dur := range durs {
			if elapsed >= dur && ms.lastStillness < dur {
				ms.eventSink.Emit("stillness", status)
				ms.lastStillness = elapsed
			}
		}
	}
	return status, nil
}

func (ms *MotionSensor) Monitor(interval time.Duration) func() {
	return Monitor(ms, interval)
}

func (ms *MotionSensor) HandleRead(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return ms.Check()
}
