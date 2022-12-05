package api

import (
	//"bytes"
	//"encoding/json"
	"net/http"
	//"os/exec"
	//"path/filepath"
	"sync"
	"time"

	"github.com/rclancey/events"
	//"github.com/rclancey/httpserver/v2"
	"github.com/rclancey/sensors/tsl2591"
	"github.com/rclancey/sensors/types"
)

type BrightnessSensor struct {
	cfg *Config
	sensor *tsl2591.TSL2591
	eventSink events.EventSink
	stop chan bool
	lock *sync.Mutex
	lastReading *types.BrightnessReading
}

func NewBrightnessSensor(cfg *Config, eventSink events.EventSink) (*BrightnessSensor, error) {
	sink := events.NewPrefixedEventSource("brightness", eventSink)
	sensor, err := tsl2591.New()
	if err != nil {
		return nil, err
	}
	bright := &BrightnessSensor{
		cfg: cfg,
		sensor: sensor,
		eventSink: sink,
		lock: &sync.Mutex{},
	}
	bright.registerEventTypes()
	return bright, nil
}

func (bright *BrightnessSensor) registerEventTypes() {
	bright.eventSink.RegisterEventType(events.NewEvent("measurement", &types.BrightnessReading{
		&tsl2591.SensorData{
			Lux: 891,
			Infrared: 283,
			Visible: 5439688,
			FullSpectrum: 5439771,
		},
		time.Now().In(time.UTC),
	}))
}

func (bright *BrightnessSensor) Check() (interface{}, error) {
	reading, err := bright.sensor.ReadSensorData()
	if err != nil {
		return nil, err
	}
	bright.lastReading = &types.BrightnessReading{reading, time.Now().In(time.UTC)}
	bright.eventSink.Emit("measurement", bright.lastReading)
	return bright.lastReading, nil
}

func (bright *BrightnessSensor) Monitor(interval time.Duration) func() {
	return Monitor(bright, interval)
}

func (bright *BrightnessSensor) HandleRead(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return bright.lastReading, nil
}
