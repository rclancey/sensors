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
	"github.com/rclancey/openweathermap"
)

type Weather struct {
	cfg *Config
	client *openweathermap.OpenWeatherMapClient
	lat float64
	lon float64
	webhooks *ThresholdWebhookList
	stop chan bool
	lock *sync.Mutex
	lastReading *openweathermap.Forecast
}

func NewWeather(cfg *Config) (*Weather, error) {
	fn, err := cfg.Abs("weather-webhooks.json")
	if err != nil {
		return nil, err
	}
	webhooks, err := NewThresholdWebhookList(fn)
	if err != nil {
		return nil, err
	}
	cacheDir := cfg.Abs("var/cache/openweathermap.org")
	err = httpserver.EnsureDir(cacheDir)
	if err != nil {
		return nil, err
	}
	cacheTime := time.Minute
	client, err := openweathermap.NewOpenWeatherMapClient(cfg.OpenWeatherMapAPIKey, cacheDir, cacheTime)
	if err != nil {
		return nil, err
	}
	return &Weather{
		cfg: cfg,
		client: client,
		webhooks: webhooks,
		lock: &sync.Mutex{},
	}, nil
}

func (weather *Weather) SetLocation(lat, lon float64) {
	if lat != weather.lat || lon != weather.lon {
		weather.lat = lat
		weather.lon = lon
		weather.lastReading = nil
		weather.Check()
	}
}

type ValueWithUnits struct {
	Value float64
	Units string
}

func (v *ValueWithUnits) GetValue() float64 {
	return v.Value
}

func (weather *Weather) Check() (float64, interface{}, error) {
	if weather.lat == 0 && weather.lon == 0 {
		return 0, nil, nil
	}
	forecast, err := weather.client.Forecast(weather.lat, weather.lon)
	if err != nil {
		return 0, nil, err
	}
	prev := weather.lastReading
	weather.lastReading = forecast
	weather.EmitEvent("forecast", forecast)
	weather.EmitEvent("pressure", &ValueWithUnits{forecast.Current.PressureHPA, "hPa"})
	weather.EmitEvent("humidity", &ValueWithUnits{forecast.Current.HumidityPct, "%"})
	weather.EmitEvent("dewpoint", &ValueWithUnits{forecast.Current.DewPointK, "K"})
	weather.EmitEvent("clouds", &ValueWithUnits{forecast.Current.CloudPct, "%"})
	weather.EmitEvent("temperature", &ValueWithUnits{forecast.Current.TempK, "K"})
	weather.EmitEvent("rain", &ValueWithUnits{forecast.Current.Rain.Last3HoursMm / 3, "mm/h"})
	if len(forecast.Daily) > 0 {
		daily := forecast.Daily[0]
		weather.EmitEvent("low", &ValueWithUnits{daily.Temp.MinK, "K"})
		weather.EmitEvent("high", &ValueWithUnits{daily.Temp.MaxK, "K"})
		weather.EmitEvent("precipitation", &ValueWithUnits{daily.PrecipitationPct, "%"})
	}
	if prev != nil && prev.Current != nil && forecast.Current != nil {
		if prev.Current.Rain.LastHourMm > 0 && forecast.Current.Rain.LastHourMm == 0 {
			weather.EmitEvent("rain-status", "stopped")
		} else if prev.Current.Rain.LastHourMm == 0 && forecast.Current.Rain.LastHourMm > 0 {
			weather.EmitEvent("rain-status", "started")
		}
	}
	for _, alert := range forecast.Alerts {
		weather.EmitEvent("alert", alert)
	}
}

func (weather *Weather) Monitor(interval time.Duration) {
	go weather.webhooks.Monitor(weather, interval)
}

func (weather *Weather) HandleRead(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return weather.lastReading, nil
}

func (weather *Weather) HandleListWebhooks(w  http.ResponseWriter, r *http.Request) (interface{}, error) {
	return weather.webhooks, nil
}

func (weather *Weather) HandleAddWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	webhook := &ThresholdWebhook{}
	err := httpserver.ReadJSON(r, webhook)
	if err != nil {
		return nil, err
	}
	weather.webhooks.RegisterWebhook(webhook)
	return weather.webhooks.List(), nil
}

func (weather *Weather) HandleRemoveWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	webhook := &ThresholdWebhook{}
	err := httpserver.ReadJSON(r, webhook)
	if err != nil {
		return nil, err
	}
	weather.webhooks.UnregisterWebhook(webhook)
	return weather.webhooks.List(), nil
}

