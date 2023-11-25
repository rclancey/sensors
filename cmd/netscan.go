package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/rclancey/sensors/netscan"
)

func main() {
	mdnsHosts := netscan.MDNSHosts()
	ch := netscan.ARP()
	arpHosts := []*netscan.HostInfo{}
	wg := &sync.WaitGroup{}
	for {
		host, ok := <-ch
		if !ok {
			break
		}
		if host.IPv4 != "" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				nmHosts, err := netscan.NMAP(host.IPv4)
				if err == nil && len(nmHosts) > 0 {
					host.OS = nmHosts[0].OS
					host.Services = nmHosts[0].Services
				} else {
					log.Println("name", host.IPv4, "error:", err)
				}
			}()
		}
		arpHosts = append(arpHosts, host)
	}
	wg.Wait()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(map[string][]*netscan.HostInfo{"mdns": mdnsHosts, "arp": arpHosts})
}
