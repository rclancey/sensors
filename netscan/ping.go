package netscan

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/go-ping/ping"
)

func PingScan() (chan *HostInfo, error) {
	ch := make(chan *HostInfo, 10)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	ips := []net.IP{}
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if ok {
			log.Println("ipnet", ipnet)
			n := len(ipnet.IP)
			if ipnet.IP[n-4] == 192 && ipnet.IP[n-3] == 168 {
				for i := 1; i < 255; i++ {
					ip := make(net.IP, n)
					copy(ip, ipnet.IP)
					ip[n-1] = byte(i)
					log.Println("will ping", ip)
					ips = append(ips, ip)
				}
			} else {
				log.Println("skipping", []byte(ipnet.IP))
			}
		} else {
			log.Printf("skipping %T %s", addr, addr)
		}
	}
	go func() {
		wg := &sync.WaitGroup{}
		for _, ip := range ips {
			xip := ip.String()
			wg.Add(1)
			go func() {
				defer wg.Done()
				pinger, err := ping.NewPinger(xip)
				if err != nil {
					log.Println("error setting up pinger for", xip, err)
					return
				}
				pinger.Count = 1
				pinger.Timeout = time.Second
				pinger.OnRecv = func(packet *ping.Packet) {
					host := NewHostInfo()
					host.IPv4 = xip
					ch <- host
				}
				pinger.Run()
			}()
		}
		wg.Wait()
		close(ch)
	}()
	return ch, nil
}
