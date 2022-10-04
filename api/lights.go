package api

import (
	"errors"
	"log"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/rclancey/events"
	"github.com/rclancey/httpserver/v2"
	"github.com/rclancey/kasa"

)

type device struct {
	kasa.SmartDevice
	State int
	LastUpdate time.Time
}

type LightStatus struct {
	cfg *Config
	eventSink events.EventSink
	devices map[string]*device
	lock *sync.Mutex
	lastUpdate time.Time
}

func NewLightStatus(cfg *Config, eventSink events.EventSink) (*LightStatus, error) {
	sink := events.NewPrefixedEventSource("lights", eventSink)
	ls := &LightStatus{
		cfg: cfg,
		eventSink: sink,
		devices: map[string]*device{},
		lock: &sync.Mutex{},
	}
	ls.registerEventTypes()
	return ls, nil
}

func (ls *LightStatus) registerEventTypes() {
	ls.eventSink.RegisterEventType(events.NewEvent("add", &device{}))
	ls.eventSink.RegisterEventType(events.NewEvent("on", &device{}))
	ls.eventSink.RegisterEventType(events.NewEvent("off", &device{}))
	ls.eventSink.RegisterEventType(events.NewEvent("lost", &device{}))
}

func (ls *LightStatus) update(dev kasa.SmartDevice) {
	xdev := &device{dev, 0, time.Now()}
	if dev.IsOn() {
		bulb, ok := dev.(*kasa.SmartBulb)
		if ok {
			xdev.State = bulb.GetLightState().Brightness
		} else {
			xdev.State = 100
		}
	} else {
		xdev.State = 0
	}
	ls.lock.Lock()
	orig := ls.devices[dev.Alias()]
	ls.devices[xdev.Alias()] = xdev
	ls.lock.Unlock()
	if orig == nil {
		ls.eventSink.Emit("add", xdev)
	} else if orig.State != 0 && xdev.State == 0 {
		ls.eventSink.Emit("off", xdev)
	} else if orig.State <= 0 && xdev.State > 0 {
		ls.eventSink.Emit("on", xdev)
	}
}

func (ls *LightStatus) lose(alias string) {
	ls.lock.Lock()
	orig := ls.devices[alias]
	ls.lock.Unlock()
	if orig != nil {
		if orig.State >= 0 {
			orig.State = -1
			ls.eventSink.Emit("lost", orig)
		}
	}
}

func (ls *LightStatus) ips() map[string]string {
	ls.lock.Lock()
	defer ls.lock.Unlock()
	ips := make(map[string]string, len(ls.devices))
	for k, dev := range ls.devices {
		ips[k] = dev.IP()
	}
	return ips
}

func (ls *LightStatus) Discover() func() {
	quitch := make(chan bool, 2)
	quit := false
	quitFunc := func() {
		if !quit {
			quit = true
			quitch <- true
		}
	}
	go func() {
		ch, err := kasa.DiscoverStream(10 * time.Second, quitch)
		if err != nil {
			log.Fatal("error setting up kasa discovery:", err)
		}
		for {
			dev, ok := <-ch
			if !ok {
				break
			}
			ls.update(dev)
		}
	}()
	return quitFunc
}

func (ls *LightStatus) Check() (interface{}, error) {
	toUpdate := ls.ips()
	wg := &sync.WaitGroup{}
	for alias, ip := range toUpdate {
		wg.Add(1)
		xalias := alias
		xip := ip
		go func() {
			defer wg.Done()
			dev, err := kasa.NewDevice(xip)
			if err == nil {
				ls.update(dev)
			} else {
				ls.lose(xalias)
			}
		}()
	}
	wg.Wait()
	ls.lock.Lock()
	defer ls.lock.Unlock()
	devs := make([]kasa.SmartDevice, len(ls.devices))
	i := 0
	for _, dev := range ls.devices {
		devs[i] = dev
		i += 1
	}
	ls.lastUpdate = time.Now()
	return devs, nil
}

func (ls *LightStatus) Monitor(interval time.Duration) func() {
	discQuit := ls.Discover()
	monQuit := Monitor(ls, interval)
	quitFunc := func() {
		discQuit()
		monQuit()
	}
	return quitFunc
}

func (ls *LightStatus) HandleRead(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ls.lock.Lock()
	out := map[string]interface{}{
		"now": ls.lastUpdate,
	}
	devs := map[string]int{}
	out["devices"] = devs
	raw := map[string]kasa.SmartDevice{}
	out["raw"] = raw
	stale := time.Now().Add(-5 * time.Minute)
	for k, v := range ls.devices {
		raw[k] = v
		if v.LastUpdate.Before(stale) {
			devs[k] = -1
		} else if v.IsOn() {
			bulb, ok := v.SmartDevice.(*kasa.SmartBulb)
			if ok {
				devs[k] = bulb.GetLightState().Brightness
			} else {
				devs[k] = 100
			}
		} else {
			devs[k] = 0
		}
	}
	defer ls.lock.Unlock()
	return out, nil
}

func (ls *LightStatus) HandlePut(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	name := path.Base(r.URL.Path)
	val, err := strconv.Atoi(r.URL.Query().Get("value"))
	if err != nil {
		return nil, httpserver.BadRequest.Wrap(err, "invalid value")
	}
	device, ok := ls.devices[name]
	if !ok {
		return nil, httpserver.NotFound
	}
	if val == 0 {
		switch dev := device.SmartDevice.(type) {
		case kasa.Switch:
			err = dev.TurnOff()
		case kasa.Dimmer:
			err = dev.SetBrightness(0)
		default:
			err = errors.New("unsouported device")
		}
	} else {
		switch dev := device.SmartDevice.(type) {
		case kasa.Dimmer:
			err = dev.SetBrightness(val)
		case kasa.Switch:
			err = dev.TurnOn()
		default:
			err = errors.New("unsupported device")
		}
	}
	if err != nil {
		return nil, err
	}
	update, err := kasa.NewDevice(device.IP())
	if err != nil {
		ls.update(update)
	}
	return device, nil
}
