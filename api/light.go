package api

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
	//"os/exec"
	//"path/filepath"
	"time"

	"github.com/rclancey/httpserver/v2"
	"github.com/rclancey/logging"
	"github.com/rclancey/sensors/tsl2591"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
)

type LightSensor struct {
	cfg *httpserver.ServerConfig
	sensor *tsl2591.TSL2591
	webhooks *ThresholdWebhookList
	stop chan bool
	lock *sync.Mutex
}

func NewLightSensor(cfg *httpserver.ServerConfig) (*LightSensor, error) {
	sensor, err := tsl2591.New()
	if err != nil {
		return nil, err
	}
	fn, err := cfg.Abs("light-sensor-webhooks.json")
	if err != nil {
		return nil, err
	}
	webhooks, err := NewThresholdWebhookList(fn)
	if err != nil {
		return nil, err
	}
	return &LightSensor{
		cfg: cfg,
		sensor: sensor,
		webhooks: webhooks,
		lock: &sync.Mutex{},
	}, nil
}

func (light *LightSensor) Check() (float64, interface{}, error) {
	reading, err := light.sensor.ReadSensorData()
	if err != nil {
		return 0, nil, err
	}
	return float64(reading.Lux), reading, nil
}

func (light *LightSensor) Monitor(interval time.Duration) {
	go light.webhooks.Monitor(light, interval)
}

func (light *LightSensor) HandleRead(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return light.sensor.ReadSensorData()
}

func (light *LightSensor) HandleAddWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	webhook := &ThresholdWebhook{}
	err := httpserver.ReadJSON(r, webhook)
	if err != nil {
		return nil, err
	}
	light.webhooks.RegisterWebhook(webhook)
	return light.webhooks.List(), nil
}

func (light *LightSensor) HandleRemoveWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	webhook := &ThresholdWebhook{}
	err := httpserver.ReadJSON(r, webhook)
	if err != nil {
		return nil, err
	}
	light.webhooks.UnregisterWebhook(webhook)
	return light.webhooks.List(), nil
}
