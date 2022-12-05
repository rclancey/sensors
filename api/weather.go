package api

import (
	//"bytes"
	//"encoding/json"
	"log"
	"net/http"
	//"os/exec"
	//"path/filepath"
	"sync"
	"time"

	"github.com/rclancey/events"
	"github.com/rclancey/httpserver/v2"
	"github.com/rclancey/openweathermap"
)

type Weather struct {
	cfg *Config
	client *openweathermap.OpenWeatherMapClient
	lat float64
	lon float64
	eventSink events.EventSink
	stop chan bool
	lock *sync.Mutex
	lastReading *openweathermap.Forecast
}

func NewWeather(cfg *Config, eventSink events.EventSink) (*Weather, error) {
	cacheDir, err := cfg.Abs("var/cache/openweathermap.org")
	if err != nil {
		return nil, err
	}
	err = httpserver.EnsureDir(cacheDir)
	if err != nil {
		return nil, err
	}
	cacheTime := time.Minute
	client, err := openweathermap.NewOpenWeatherMapClient(cfg.OpenWeatherMapAPIKey, cacheDir, cacheTime)
	if err != nil {
		return nil, err
	}
	weather := &Weather{
		cfg: cfg,
		client: client,
		eventSink: events.NewPrefixedEventSource("weather", eventSink),
		lat: cfg.Location.Latitude,
		lon: cfg.Location.Longitude,
		lock: &sync.Mutex{},
	}
	weather.registerEventTypes()
	return weather, nil
}

func (w *Weather) registerEventTypes() {
	forecast, err := w.client.Forecast(w.lat, w.lon)
	if err != nil {
		log.Println("error getting forecast:", err)
		return
	}
	w.lastReading = forecast
	w.eventSink.RegisterEventType(events.NewEvent("location", map[string]float64{"lat": w.lat, "lon": w.lon}))
	w.eventSink.RegisterEventType(events.NewEvent("forecast", forecast))
	w.eventSink.RegisterEventType(events.NewEvent("pressure", &ValueWithUnits{"pressure", forecast.Current.PressureHPa, "hPa"}))
	w.eventSink.RegisterEventType(events.NewEvent("humidity", &ValueWithUnits{"humidity", forecast.Current.HumidityPct, "%"}))
	w.eventSink.RegisterEventType(events.NewEvent("dewpoint", &ValueWithUnits{"dew_point", forecast.Current.DewPointK, "K"}))
	w.eventSink.RegisterEventType(events.NewEvent("clouds", &ValueWithUnits{"clouds", forecast.Current.CloudPct, "%"}))
	w.eventSink.RegisterEventType(events.NewEvent("temp", &ValueWithUnits{"temp", forecast.Current.TempK, "K"}))
	w.eventSink.RegisterEventType(events.NewEvent("rain", &ValueWithUnits{"rain", forecast.Current.Rain.Last3HoursMm / 3, "mm/h"}))
	if len(forecast.Daily) > 0 {
		daily := forecast.Daily[0]
		w.eventSink.RegisterEventType(events.NewEvent("low", &ValueWithUnits{"temp_min", daily.Temp.MinK, "K"}))
		w.eventSink.RegisterEventType(events.NewEvent("high", &ValueWithUnits{"temp_max", daily.Temp.MaxK, "K"}))
		w.eventSink.RegisterEventType(events.NewEvent("pop", &ValueWithUnits{"pop", daily.PrecipitationPct, "%"}))
	}
	w.eventSink.RegisterEventType(events.NewEvent("rain-status", "stopped"))
	w.eventSink.RegisterEventType(events.NewEvent("alert", &openweathermap.Alert{}))
}

func (w *Weather) SetLocation(lat, lon float64) {
	if lat != w.lat || lon != w.lon {
		w.lat = lat
		w.lon = lon
		w.lastReading = nil
		w.eventSink.Emit("location", map[string]float64{"lat": w.lat, "lon": w.lon})
		w.Check()
	}
}

type ValueWithUnits struct {
	Name string
	Value float64
	Units string
}

func (v *ValueWithUnits) GetValue() float64 {
	return v.Value
}

func (w *Weather) Check() (interface{}, error) {
	if w.lat == 0 && w.lon == 0 {
		log.Println("no location for weather")
		return nil, nil
	}
	forecast, err := w.client.Forecast(w.lat, w.lon)
	if err != nil {
		log.Println("error getting forecast:", err)
		return nil, err
	}
	if forecast == nil {
		log.Println("no forecast!")
		return nil, nil
	}
	prev := w.lastReading
	w.lastReading = forecast
	w.eventSink.Emit("forecast", forecast)
	w.eventSink.Emit("pressure", &ValueWithUnits{"pressure", forecast.Current.PressureHPa, "hPa"})
	w.eventSink.Emit("humidity", &ValueWithUnits{"humidity", forecast.Current.HumidityPct, "%"})
	w.eventSink.Emit("dewpoint", &ValueWithUnits{"dew_point", forecast.Current.DewPointK, "K"})
	w.eventSink.Emit("clouds", &ValueWithUnits{"clouds", forecast.Current.CloudPct, "%"})
	w.eventSink.Emit("temp", &ValueWithUnits{"temp", forecast.Current.TempK, "K"})
	w.eventSink.Emit("rain", &ValueWithUnits{"rain", forecast.Current.Rain.Last3HoursMm / 3, "mm/h"})
	if len(forecast.Daily) > 0 {
		daily := forecast.Daily[0]
		w.eventSink.Emit("low", &ValueWithUnits{"temp_min", daily.Temp.MinK, "K"})
		w.eventSink.Emit("high", &ValueWithUnits{"temp_max", daily.Temp.MaxK, "K"})
		w.eventSink.Emit("pop", &ValueWithUnits{"pop", daily.PrecipitationPct, "%"})
	}
	if prev != nil {
		if prev.Current.Rain.LastHourMm > 0 && forecast.Current.Rain.LastHourMm == 0 {
			w.eventSink.Emit("rain-status", "stopped")
		} else if prev.Current.Rain.LastHourMm == 0 && forecast.Current.Rain.LastHourMm > 0 {
			w.eventSink.Emit("rain-status", "started")
		}
	}
	for _, alert := range forecast.Alerts {
		w.eventSink.Emit("alert", alert)
	}
	return forecast, nil
}

func (w *Weather) Monitor(interval time.Duration) func() {
	return Monitor(w, interval)
}

func (weather *Weather) HandleRead(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return weather.lastReading, nil
}
