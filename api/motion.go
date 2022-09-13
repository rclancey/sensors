package api

import (
	//"encoding/json"
	"fmt"
	"net/http"
	//"os/exec"
	//"path/filepath"
	"time"

	"github.com/rclancey/httpserver/v2"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
)

type MotionSensorStatus struct {
	Now time.Time `json:"now"`
	LastMotion time.Time `json:"last_motion"`
	ElapsedTime float64 `json:"elapsed_time"`
}

type MotionSensor struct {
	cfg *httpserver.ServerConfig
	line *gpiod.Line
	lastMotion time.Time
	webhooks *ThresholdWebhookList
}

func NewMotionSensor(cfg *httpserver.ServerConfig) (*MotionSensor, error) {
	chipIdx := 0
	lineId := rpi.GPIO17
	line, err := gpiod.RequestLine(fmt.Sprintf("gpiochip%d", chipIdx), lineId)
	if err != nil {
		return nil, err
	}
	fn, err := cfg.Abs("motion-sensor-webhooks.json")
	if err != nil {
		return nil, err
	}
	webhooks, err := NewThresholdWebhookList(fn)
	if err != nil {
		return nil, err
	}
	return &MotionSensor{
		cfg: cfg,
		line: line,
		webhooks: webhooks,
	}, nil
}

func (ms *MotionSensor) Check() (float64, interface{}, error) {
	val, err := ms.line.Value()
	if err != nil {
		return 0, nil, err
	}
	status := &MotionSensorStatus{
		Now: time.Now().In(time.UTC),
		LastMotion: ms.lastMotion,
		ElapsedTime: 0,
	}
	if val != 0 {
		ms.lastMotion = status.Now
		status.LastMotion = status.Now
	} else {
		status.ElapsedTime = status.Now.Sub(status.LastMotion).Seconds()
	}
	return status.ElapsedTime, status, nil
}

func (ms *MotionSensor) Monitor(interval time.Duration) {
	go ms.webhooks.Monitor(ms, interval)
}

func (ms *MotionSensor) HandleRead(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	_, data, err := ms.Check()
	return data, err
}

func (ms *MotionSensor) HandleAddWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	webhook := &ThresholdWebhook{}
	err := httpserver.ReadJSON(r, webhook)
	if err != nil {
		return nil, err
	}
	ms.webhooks.RegisterWebhook(webhook)
	return ms.webhooks.List(), nil
}

func (ms *MotionSensor) HandleRemoveWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	webhook := &ThresholdWebhook{}
	err := httpserver.ReadJSON(r, webhook)
	if err != nil {
		return nil, err
	}
	ms.webhooks.UnregisterWebhook(webhook)
	return ms.webhooks.List(), nil
}
