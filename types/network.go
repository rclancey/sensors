package types

import (
	"time"
)

type HostInfo struct {
	MAC        string    `json:"mac"`
	IPv4       string    `json:"ipv4"`
	Hostname   string    `json:"hostname"`
	Connection string    `json:"connection"`
	LastSeen   time.Time `json:"last_seen"`
	Active     bool      `json:"active"`
}

