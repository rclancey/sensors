package main

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/mdlayher/arp"
)

func main() {
	ifi, err := net.InterfaceByName("wlan0")
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	for i := 1; i < 255; i++ {
		ipstr := fmt.Sprintf("129.168.0.%d", i)
		wg.Add(1)
		go func() {
			// Set up ARP client with socket
			c, err := arp.Dial(ifi)
			if err != nil {
				log.Fatal(err)
			}
			defer c.Close()
			// Set request deadline from flag
			if err := c.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
				log.Fatal(err)
			}
			// Request hardware address for IP address
			ip, err := netip.ParseAddr(ipstr)
			if err != nil {
				log.Fatal(err)
			}
			mac, err := c.Resolve(ip)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s -> %s", ip, mac)
			wg.Done()
		}()
	}
	wg.Wait()
}
