package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}
	for _, addr := range addrs {
		fmt.Printf("%T = %s\n", addr, addr)
		ipnet, ok := addr.(*net.IPNet)
		if ok {
			fmt.Println(ipnet.Mask.Size())
		}
	}
}
