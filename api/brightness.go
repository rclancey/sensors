package api

import (
	//"bytes"
	//"encoding/json"
	"math"
	"net/http"
	//"os/exec"
	//"path/filepath"
	"sync"
	"time"

	"github.com/rclancey/events"
	"github.com/rclancey/gosolar"
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
	weather *Weather
}

func NewBrightnessSensor(cfg *Config, eventSink events.EventSink, weather *Weather) (*BrightnessSensor, error) {
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
		weather: weather,
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
		543.2,
	}))
}

func (bright *BrightnessSensor) calcBrightness() float64 {
	if bright.weather == nil || bright.weather.lastReading == nil {
		return 0
	}
	cur := bright.weather.lastReading.Current
	temp := cur.TempK
	pres := cur.PressureHPa * 100
	cloud := cur.CloudPct / 100
	when := time.Now().In(time.UTC)
	when = when.Add(90 * time.Minute) // correctiion factor
	lat := bright.cfg.Location.Latitude
	lon := bright.cfg.Location.Longitude
	elev := bright.cfg.Location.Elevation
	alt := solar.GetAltitude(lat, lon, elev, when, &temp, &pres)
	if alt < 0 {
		return 0
	}
	rad := solar.GetRadiationDirect(when, alt)
	rad *= math.Sin((alt - 10) * math.Pi / 180)
	// atten := 1.0 - (0.75 * math.Pow(cloud, 3.4))
	atten := 1.0 - (0.85 * (cloud * cloud))
	lux := rad * atten / 0.0079
	if lux < 0 {
		lux = 0
	}
	return lux
}

func (bright *BrightnessSensor) Check() (interface{}, error) {
	reading, err := bright.sensor.ReadSensorData()
	if err != nil {
		return nil, err
	}
	calc := bright.calcBrightness()
	bright.lastReading = &types.BrightnessReading{reading, time.Now().In(time.UTC), calc}
	bright.eventSink.Emit("measurement", bright.lastReading)
	Measure("brightness_measured", map[string]string{"units": "lux"}, float64(reading.Lux))
	Measure("brightness_calculated", map[string]string{"units": "lux"}, calc)
	return bright.lastReading, nil
}

func (bright *BrightnessSensor) Monitor(interval time.Duration) func() {
	return Monitor(bright, interval)
}

func (bright *BrightnessSensor) HandleRead(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return bright.lastReading, nil
}
