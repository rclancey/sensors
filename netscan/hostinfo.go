package netscan

import (
	"sync"
	"time"
)

type Service struct {
	Name string
	Port int
}

type HostInfo struct {
	MAC string `json:"mac"`
	IPv4 string `json:"ipv4"`
	IPv6 string `json:"ipv6"`
	Name string `json:"name"`
	OS string `json:"os"`
	Services []*Service `json:"services"`
	LastSeen time.Time `json:"last_seen"`
	HasARP bool `json:"has_arp"`
	HasMDNS bool `json:"has_mdns"`
	HasNMAP bool `json:"has_nmap"`
	mutex *sync.Mutex
}

func NewHostInfo() *HostInfo {
	return &HostInfo{
		LastSeen: time.Now(),
		mutex: &sync.Mutex{},
	}
}

func (host *HostInfo) HasService(name string) bool {
	host.mutex.Lock()
	defer host.mutex.Unlock()
	for _, svc := range host.Services {
		if svc.Name == name {
			return true
		}
	}
	return false
}

func (host *HostInfo) AddService(name string, port int) {
	host.mutex.Lock()
	defer host.mutex.Unlock()
	for _, svc := range host.Services {
		if svc.Port == port {
			return
		}
	}
	host.Services = append(host.Services, &Service{Name: name, Port: port})
}
