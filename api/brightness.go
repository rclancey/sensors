package api

import (
	//"bytes"
	//"encoding/json"
	"net/http"
	//"os/exec"
	//"path/filepath"
	"sync"
	"time"

	"github.com/rclancey/httpserver/v2"
	"github.com/rclancey/sensors/tsl2591"
)

type BrightnessReading struct {
	*tsl2591.SensorData
	Now time.Time `json:"now"`
}

func (br *BrightnessReading) Value() float64 {
	return float64(br.Lux)
}

type BrightnessSensor struct {
	cfg *Config
	sensor *tsl2591.TSL2591
	eventSink events.EventSink
	stop chan bool
	lock *sync.Mutex
	lastReading *BrightnessReading
}

func NewBrightnessSensor(cfg *Config, eventSink events.EventSink) (*BrightnessSensor, error) {
	sink := events.NewPrefixedEventSource("brightness", eventSink)
	sensor, err := tsl2591.New()
	if err != nil {
		return nil, err
	}
	fn, err := cfg.Abs("brightness-sensor-webhooks.json")
	if err != nil {
		return nil, err
	}
	webhooks, err := NewThresholdWebhookList(fn)
	if err != nil {
		return nil, err
	}
	return &BrightnessSensor{
		cfg: cfg,
		sensor: sensor,
		eventSink: sink,
		lock: &sync.Mutex{},
	}, nil
}

func (bright *BrightnessSensor) Check() (interface{}, error) {
	reading, err := bright.sensor.ReadSensorData()
	if err != nil {
		return 0, nil, err
	}
	bright.lastReading = &BrightnessReading{reading, time.Now()}
	bright.eventSink.Emit("measurement", bright.lastReading)
	return bright.lastReading, nil
}

func (bright *BrightnessSensor) Monitor(interval time.Duration) func() {
	return Monitor(bright, interval)
}

func (bright *BrightnessSensor) HandleRead(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return bright.lastReading, nil
}
