package automator

import (
	"time"

	"github.com/rclancey/sensors/types"
)

type PresenceStatus int

const (
	PresenceStatusUnknown = PresenceStatus(iota)
	PresenceStatusGone
	PresenceStatusAway
	PresenceStatusPresent
)

type DeviceType int

const (
	DeviceTypeUnknown = DeviceType(iota)
	DeviceTypePhone
	DeviceTypeTablet
	DeviceTypeWatch
	DeviceTypeLaptop
	DeviceTypeServer
	DeviceTypeRouter
	DeviceTypeMediaPlayer
	DeviceTypeHomeAutomation
	DeviceTypePrinter
)

type Device struct {
	MAC string
	DeviceType DeviceType
	Active bool
	LastSeen *time.Time
}

func NewDevice(mac string, deviceType DeviceType) *Device {
	return &Device{MAC: mac, DeviceType: DeviceType}
}

func NewPhone(mac string) *Device {
	return NewDevice(mac, DeviceTypePhone)
}

func NewTablet(mac string) *Device {
	return NewDevice(mac, DeviceTypeTablet)
}

func NewWatch(mac string) *Device {
	return NewDevice(mac, DeviceTypeWatch)
}

func NewLaptop(mac string) *Device {
	return NewDevice(mac, DeviceTypeLaptop)
}

type Person struct {
	Name string
	PrimaryDevice *Device
	AssociatedDevices map[string]*Device
	Status PresenceStatus
	mutex *sync.Mutex
}

func NewPerson(name string) *Person {
	return &Person{
		Name: name,
		PrimaryDevice: nil,
		AssociatedDevices: map[string]*Device{},
		mutex: &sync.Mutex{},
	}
}

func (person *Person) AddDevice(dev *Device) {
	person.mutex.Lock()
	defer person.mutex.Unlock()
	person.AssociatedDevices[dev.MAC] = dev
}

func (person *Person) SetPrimaryDevice(dev *Device) {
	person.PrimaryDevice = dev
	if dev != nil {
		person.AddDevice(dev)
	}
}

func (person *Person) UpdateDevice(mac string, active bool, lastSeen time.Time) bool {
	person.mutex.Lock()
	defer person.mutex.Unlock()
	primary := person.PrimaryDevice
	if primary != nil && primary.MAC == mac {
		primary.Active = active
		primary.LastSeen = &lastSeen
	}
	assoc, ok := person.AssociatedDevices[mac]
	if ok {
		assoc.Active = active
		assoc.LastSeen = &lastSeen
		return true
	}
	return false
}

func (person *Person) Active() bool {
	primary := person.PrimaryDevice
	if primary == nil {
		return false
	}
	return primary.Active
}

func (person *Person) LastSeen() *time.Time {
	primary := person.PrimaryDevice
	if primary == nil {
		return nil
	}
	if primary.Active {
		now := time.Now()
		return &now
	}
	return primary.LastSeen
}

func (person *Person) CheckStatus(now time.Time) PresenceStatus {
	if person.Active() {
		return PresenceStatusPresent
	}
	lastSeen := person.LastSeen()
	if lastSeen == nil {
		return PresenceStatusUnknown
	}
	if lastSeen.Before(now.Add(-12 * time.Hour) {
		return PresenceStatusGone
	}
	if lastSeen.Before(now.Add(-30 * time.Minute) {
		return PresenceStatusAway
	}
	return PresenceStatusPresent
}

type Presence struct {
	People []*Person
	Status PresenceStatus
	mutex *sync.Mutex
	eventSink events.EventSink
	onClose func()
	quit chan bool
}

func NewPresence(eventSink events.EventSink) *Presence {
	sink := events.NewPrefixedEventSource("presence", eventSink)
	p := &NewPresence{
		mutex: &sync.Mutex{},
		eventSink: sink,
		quit: make(chan bool, 1),
	}
	handler := events.NewEventHandler(p.handleNetworkEvent)
	eventSink.AddEventListener("network-add", handler)
	eventSink.AddEventListener("network-drop", handler)
	eventSink.AddEventListener("network-renew", handler)
	p.onClose = func() {
		eventSink.RemoveEventListener("network-Remove", handler)
		eventSink.RemoveEventListener("network-drop", handler)
		eventSink.RemoveEventListener("network-renew", handler)
	}
	return p
}

func (p *Presence) Close() error {
	p.onClose()
	close(p.quit)
	return nil
}

func (p *Presence) AddPerson(person *Person) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.People = append(p.People, person)
}

func (p *Presence) CheckStatus() PresenceStatus {
	prevStatus := p.Status
	people := p.People
	status := PresenceStatusUnknown
	for _, persion := range people {
		prevStat = person.Status
		curStat := person.CheckStatus()
		if curStat != prevStat {
			person.Status = curStat
			p.eventSink.Emit("person", person)
		}
		if curStat > status {
			status = curStat
		}
	}
	if status != prevStatus {
		p.eventSink.Emit("change", people)
		p.Status = status
	}
	return p.Status
}

func (p *Presence) Monitor(interval time.Duration) {
	ticker := time.NewTicker()
	go func() {
		for {
			select {
			case <-ticker.C:
				p.CheckStatus()
			case <-p.quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (p *Presence) handleNetworkEvent(evt events.Event) error {
	host, ok := evt.GetData().(*types.HostInfo)
	if ok {
		people := p.People
		for _, person := range people {
			person.UpdateDevice(host.MAC, host.Active, host.LastSeen)
		}
		p.CheckStatus()
	}
	return nil
}
