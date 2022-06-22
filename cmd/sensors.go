package main

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

var cfg *httpserver.ServerConfig

func main() {
	var err error
	cfg, err = httpserver.Configure()
	if err != nil {
		log.Fatal(err)
	}
	errlog, err := cfg.Logging.ErrorLogger()
	if err != nil {
		log.Fatal(err)
	}
	errlog.Colorize()
	errlog.SetLevelColor(logging.INFO, logging.ColorCyan, logging.ColorDefault, logging.FontDefault)
	errlog.SetLevelColor(logging.LOG, logging.ColorMagenta, logging.ColorDefault, logging.FontDefault)
	errlog.SetLevelColor(logging.TRACE, logging.ColorYellow, logging.ColorDefault, logging.FontDefault)
	errlog.SetLevelColor(logging.NONE, logging.ColorHotPink, logging.ColorDefault, logging.FontDefault)
	errlog.SetTimeFormat("2006-01-02 15:04:05.000")
	errlog.SetTimeColor(logging.ColorDefault, logging.ColorDefault, logging.FontItalic|logging.FontLight)
	errlog.SetSourceFormat("%{basepath}:%{function}:%{linenumber}:")
	errlog.SetSourceColor(logging.ColorGreen, logging.ColorDefault, logging.FontDefault)
	errlog.SetPrefixColor(logging.ColorOrange, logging.ColorDefault, logging.FontDefault)
	errlog.SetMessageColor(logging.ColorDefault, logging.ColorDefault, logging.FontDefault)
	errlog.MakeDefault()

	ms, err := NewMotionSensor(0, rpi.GPIO17)
	if err != nil {
		log.Fatal(err)
	}

	srv, err := httpserver.NewServer(cfg)
	if err != nil {
		errlog.Fatalln("can't create server:", err)
	}

	srv.GET("/light", httpserver.HandlerFunc(LightSensorHandler))
	srv.GET("/motion", httpserver.HandlerFunc(ms.Handler))
	errlog.Infoln("API server ready")
	srv.ListenAndServe()
	errlog.Infoln("API server shutting down")
}

func LightSensorHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	sensor, err := tsl2591.New()
	if err != nil {
		return nil, err
	}
	return sensor.ReadSensorData()
	/*
	cmd := exec.Command(filepath.Join(cfg.ServerRoot, "light_sensor", "light-sensor"))
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	errBytes := stderr.Bytes()
	if len(errBytes) > 0 {
		log.Println(string(errBytes))
	}
	if err != nil {
		return nil, err
	}
	obj := map[string]interface{}{}
	err = json.Unmarshal(stdout.Bytes(), &obj)
	if err != nil {
		log.Printf("error unmarshaling %s: %s", string(stdout.Bytes()), err)
		return nil, err
	}
	return obj, nil
	*/
}

type MotionSensor struct {
	line *gpiod.Line
	lastMotion time.Time
	stop chan bool
}

func NewMotionSensor(chipIdx, lineId int) (*MotionSensor, error) {
	line, err := gpiod.RequestLine(fmt.Sprintf("gpiochip%d", chipIdx), lineId)
	if err != nil {
		return nil, err
	}
	sensor := &MotionSensor{line: line, stop: make(chan bool)}
	go sensor.Monitor()
	return sensor, nil
}

func (ms *MotionSensor) Monitor() {
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			val, err := ms.line.Value()
			if err == nil && val != 0 {
				ms.lastMotion = time.Now()
			}
		case <-ms.stop:
			ticker.Stop()
			return
		}
	}
}

func (ms *MotionSensor) Stop() {
	ms.stop <- true
}

type MotionSensorStatus struct {
	Now time.Time `json:"now"`
	LastMotion time.Time `json:"last_motion"`
	ElapsedTime float64 `json:"elapsed_time"`
}

func (ms *MotionSensor) Handler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	now := time.Now().In(time.UTC)
	return &MotionSensorStatus{
		Now: now,
		LastMotion: ms.lastMotion.In(time.UTC),
		ElapsedTime: now.Sub(ms.lastMotion).Seconds(),
	}, nil
}
