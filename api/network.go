package api

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rclancey/events"
	//"github.com/rclancey/httpserver/v2"
	"github.com/rclancey/archer-ax50"
)

type HostInfo struct {
	MAC        string    `json:"mac"`
	IPv4       string    `json:"ipv4"`
	Hostname   string    `json:"hostname"`
	Connection string    `json:"connection"`
	LastSeen   time.Time `json:"last_seen"`
	Active     bool      `json:"active"`
}

type NetworkStatusResponse struct {
	LastUpdate time.Time   `json:"now"`
	Hosts      []*HostInfo `json:"hosts"`
	Active     int         `json:"active"`
}

type NetworkStatus struct {
	lastUpdate time.Time `json:"now"`
	hosts map[string]*HostInfo `json:"hosts"`
	active int `json:"active"`
	cfg *Config
	eventSink events.EventSink
	mutex *sync.Mutex
}

func NewNetworkStatus(cfg *Config, eventSink events.EventSink) (*NetworkStatus, error) {
	ns := &NetworkStatus{
		lastUpdate: time.Now(),
		hosts: map[string]*HostInfo{},
		cfg: cfg,
		eventSink: events.NewPrefixedEventSource("network", eventSink),
		mutex: &sync.Mutex{},
	}
	ns.registerEventTypes()
	return ns, nil
}

func (ns *NetworkStatus) registerEventTypes() {
	now := time.Now()
	ns.eventSink.RegisterEventType(events.NewEvent("add", &HostInfo{
		MAC: "40:3f:8c:72:bd:a4",
		IPv4: "129.168.0.100",
		Hostname: "laptop",
		Connection: "2.4G",
		LastSeen: now,
		Active: true,
	}))
	ns.eventSink.RegisterEventType(events.NewEvent("drop", &HostInfo{
		MAC: "40:3f:8c:72:bd:a4",
		IPv4: "129.168.0.100",
		Hostname: "laptop",
		Connection: "2.4G",
		LastSeen: now.Add(-5 * time.Minute),
		Active: false,
	}))
	ns.eventSink.RegisterEventType(events.NewEvent("change", &HostInfo{
		MAC: "40:3f:8c:72:bd:a4",
		IPv4: "129.168.0.101",
		Hostname: "laptop",
		Connection: "2.4G",
		LastSeen: now,
		Active: true,
	}))
}

func (ns *NetworkStatus) Check() (interface{}, error) {
	client, err := ax50.NewClient(ns.cfg.Router.IP)
	if err != nil {
		return nil, err
	}
	err = client.Login(ns.cfg.Router.Password)
	if err != nil {
		return nil, err
	}
	defer client.Logout()
	clients, err := client.GetClientList()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	allClients := append(clients.Wireless, clients.Wired...)
	ns.mutex.Lock()
	defer ns.mutex.Unlock()
	ns.lastUpdate = now
	ns.active = len(allClients)
	missing := map[string]bool{}
	for k := range ns.hosts {
		missing[k] = true
	}
	changes := 0
	for _, c := range allClients {
		c.MAC = strings.ReplaceAll(strings.ToLower(c.MAC), "-", ":")
		host := ns.hosts[c.MAC]
		if host == nil {
			host = &HostInfo{
				MAC: c.MAC,
				IPv4: c.IPv4,
				Hostname: c.Hostname,
				Connection: c.WireType,
				LastSeen: now,
				Active: true,
			}
			hostname := ns.cfg.Router.Hosts[c.MAC]
			if hostname != "" {
				host.Hostname = hostname
			}
			ns.hosts[c.MAC] = host
			changes += 1
			ns.eventSink.Emit("add", host)
		} else {
			delete(missing, c.MAC)
			changed := false
			if host.IPv4 != c.IPv4 {
				host.IPv4 = c.IPv4
				changed = true
			}
			hostname := ns.cfg.Router.Hosts[c.MAC]
			if hostname == "" {
				hostname = c.Hostname
			}
			if host.Hostname != hostname {
				host.Hostname = hostname
				changed = true
			}
			if host.Connection != c.WireType {
				host.Connection = c.WireType
				changed = true
			}
			host.LastSeen = now
			if !host.Active {
				host.Active = true
				ns.eventSink.Emit("add", host)
			} else if changed {
				ns.eventSink.Emit("change", host)
			}
		}
	}
	for k := range missing {
		host := ns.hosts[k]
		if host.Active {
			host.Active = false
			ns.eventSink.Emit("drop", host)
		}
	}
	resp := ns.makeResponse()
	return resp, nil
}

func (ns *NetworkStatus) makeResponse() *NetworkStatusResponse {
	resp := &NetworkStatusResponse{
		LastUpdate: ns.lastUpdate,
		Hosts: make([]*HostInfo, len(ns.hosts)),
		Active: ns.active,
	}
	i := 0
	for _, v := range ns.hosts {
		resp.Hosts[i] = v
		i += 1
	}
	sort.Slice(resp.Hosts, func(i, j int) bool {
		if resp.Hosts[i].IPv4 == resp.Hosts[j].IPv4 {
			return resp.Hosts[i].MAC < resp.Hosts[j].MAC
		}
		return resp.Hosts[i].IPv4 < resp.Hosts[j].IPv4
	})
	return resp
}

func (ns *NetworkStatus) Monitor(interval time.Duration) func() {
	return Monitor(ns, interval)
}

func (ns *NetworkStatus) HandleRead(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()
	resp := ns.makeResponse()
	return resp, nil
}
