package netscan

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

var ipreArp = regexp.MustCompile(`\((\d+\.\d+\.\d+\.\d+)\)`)
var macreArp = regexp.MustCompile(` at (.*) on `)

func readArp(cmd *exec.Cmd) chan *HostInfo {
	ch := make(chan *HostInfo, 10)
	go func() {
		defer close(ch)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Println("error getting stdout pipe:", err)
			return
		}
		err = cmd.Start()
		if err != nil {
			log.Println("error running command:", err)
			return
		}
		bufOut := bufio.NewReader(stdout)
		for {
			line, err := bufOut.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				log.Fatalln("error reading from command:", err)
				cmd.Process.Kill()
			}
			ipmatch := ipreArp.FindStringSubmatch(line)
			if ipmatch == nil {
				continue
			}
			macmatch := macreArp.FindStringSubmatch(line)
			if macmatch == nil {
				continue
			}
			ip := ipmatch[1]
			mac := strings.Split(macmatch[1], " ")[0]
			if mac == "<incomplete>" {
				continue
			}
			host := NewHostInfo()
			host.MAC = mac
			host.IPv4 = ip
			host.HasARP = true
			ch <- host
		}
		cmd.Wait()
	}()
	return ch
}

func ARPClear() {
	cmd := exec.Command("sudo", "ip", "-s", "-s", "neigh", "flush", "all")
	cmd.Run()
}

func ARP(ip string) (*HostInfo, error) {
	cmd := exec.Command("arp", "-n", ip)
	ch := readArp(cmd)
	ok := true
	var host *HostInfo
	for ok {
		host, ok = <-ch
	}
	return host, nil
}

func ARPScan() chan *HostInfo {
	cmd := exec.Command("arp", "-an")
	return readArp(cmd)
}

/*
pi@raspberrypi:~/src/sensors $ arp -an
? (192.168.0.199) at <incomplete> on wlan0
? (192.168.0.50) at <incomplete> on wlan0
? (192.168.0.238) at be:60:c2:0a:90:89 [ether] on wlan0
? (192.168.0.191) at <incomplete> on wlan0
? (192.168.0.12) at <incomplete> on wlan0
? (192.168.0.221) at b2:b7:47:9d:6a:08 [ether] on wlan0
? (192.168.0.123) at bc:d0:74:26:66:3e [ether] on wlan0
? (192.168.0.200) at <incomplete> on wlan0
? (192.168.0.153) at d0:81:7a:ad:54:26 [ether] on wlan0
? (192.168.0.55) at <incomplete> on wlan0
? (192.168.0.243) at c6:65:e9:67:b0:2c [ether] on wlan0
? (192.168.0.64) at 00:24:36:ec:7a:16 [ether] on wlan0
? (192.168.0.205) at <incomplete> on wlan0
? (192.168.0.154) at <incomplete> on wlan0
? (192.168.0.69) at <incomplete> on wlan0
? (192.168.0.227) at 10:1c:0c:36:91:6f [ether] on wlan0
? (192.168.0.1) at 40:3f:8c:72:bd:a4 [ether] on wlan0
? (192.168.0.206) at <incomplete> on wlan0
? (192.168.0.61) at <incomplete> on wlan0
? (192.168.0.40) at <incomplete> on wlan0
? (192.168.0.70) at <incomplete> on wlan0
? (192.168.0.228) at <incomplete> on wlan0
? (192.168.0.211) at <incomplete> on wlan0
? (192.168.0.45) at <incomplete> on wlan0
? (192.168.0.75) at 10:27:f5:44:b8:27 [ether] on wlan0
? (192.168.0.182) at <incomplete> on wlan0
? (192.168.0.7) at <incomplete> on wlan0
? (192.168.0.212) at <incomplete> on wlan0
? (192.168.0.195) at <incomplete> on wlan0
? (192.168.0.46) at <incomplete> on wlan0
? (192.168.0.76) at 10:27:f5:3f:ce:c9 [ether] on wlan0
? (192.168.0.187) at <incomplete> on wlan0
? (192.168.0.8) at <incomplete> on wlan0
? (192.168.0.166) at b0:fc:0d:f4:70:5a [ether] on wlan0
? (192.168.0.119) at 00:31:92:44:c9:fd [ether] on wlan0
? (192.168.0.196) at <incomplete> on wlan0
? (192.168.0.51) at <incomplete> on wlan0
? (192.168.0.128) at <incomplete> on wlan0
? (192.168.0.81) at b0:a7:b9:6e:78:a8 [ether] on wlan0
? (192.168.0.188) at 94:9f:3e:89:86:04 [ether] on wlan0
? (192.168.0.218) at <incomplete> on wlan0
? (192.168.0.201) at <incomplete> on wlan0
? (192.168.0.52) at <incomplete> on wlan0
? (192.168.0.82) at <incomplete> on wlan0
? (192.168.0.65) at <incomplete> on wlan0
? (192.168.0.125) at <incomplete> on wlan0
? (192.168.0.202) at <incomplete> on wlan0
? (192.168.0.155) at <incomplete> on wlan0
? (192.168.0.134) at f0:24:75:7f:0c:9e [ether] on wlan0
? (192.168.0.87) at <incomplete> on wlan0
? (192.168.0.66) at <incomplete> on wlan0
? (192.168.0.126) at <incomplete> on wlan0
? (192.168.0.207) at <incomplete> on wlan0
? (192.168.0.58) at <incomplete> on wlan0
? (192.168.0.139) at cc:d2:81:70:e1:85 [ether] on wlan0
? (192.168.0.88) at <incomplete> on wlan0
? (192.168.0.41) at <incomplete> on wlan0
? (192.168.0.229) at <incomplete> on wlan0
? (192.168.0.3) at <incomplete> on wlan0
? (192.168.0.208) at <incomplete> on wlan0
? (192.168.0.63) at <incomplete> on wlan0
? (192.168.0.25) at <incomplete> on wlan0
? (192.168.0.183) at <incomplete> on wlan0
? (192.168.0.4) at <incomplete> on wlan0
? (192.168.0.213) at <incomplete> on wlan0
? (192.168.0.192) at <incomplete> on wlan0
? (192.168.0.145) at <incomplete> on wlan0
? (192.168.0.47) at <incomplete> on wlan0
? (192.168.0.77) at <incomplete> on wlan0
? (192.168.0.26) at <incomplete> on wlan0
? (192.168.0.184) at <incomplete> on wlan0
? (192.168.0.9) at <incomplete> on wlan0
? (192.168.0.214) at <incomplete> on wlan0
? (192.168.0.146) at <incomplete> on wlan0
? (192.168.0.99) at 00:11:32:98:27:33 [ether] on wlan0
? (192.168.0.48) at <incomplete> on wlan0
? (192.168.0.129) at <incomplete> on wlan0
? (192.168.0.78) at b0:a7:b9:6e:78:6c [ether] on wlan0
? (192.168.0.189) at <incomplete> on wlan0
? (192.168.0.219) at <incomplete> on wlan0
? (192.168.0.151) at 8c:85:90:b2:b0:18 [ether] on wlan0
? (192.168.0.130) at <incomplete> on wlan0
? (192.168.0.83) at <incomplete> on wlan0
? (192.168.0.190) at <incomplete> on wlan0
? (192.168.0.220) at <incomplete> on wlan0
? (192.168.0.203) at <incomplete> on wlan0
? (192.168.0.54) at <incomplete> on wlan0
? (192.168.0.67) at <incomplete> on wlan0
? (192.168.0.127) at <incomplete> on wlan0
? (192.168.0.204) at <incomplete> on wlan0
? (192.168.0.59) at <incomplete> on wlan0
? (192.168.0.247) at <incomplete> on wlan0
? (192.168.0.226) at <incomplete> on wlan0
? (192.168.0.209) at <incomplete> on wlan0
? (192.168.0.60) at 00:23:df:99:07:d4 [ether] on wlan0
? (192.168.0.141) at <incomplete> on wlan0
? (192.168.0.43) at <incomplete> on wlan0
? (192.168.0.248) at <incomplete> on wlan0
? (192.168.0.5) at <incomplete> on wlan0
? (192.168.0.210) at a8:ee:c6:00:3b:e0 [ether] on wlan0
? (192.168.0.193) at <incomplete> on wlan0
? (192.168.0.142) at <incomplete> on wlan0
? (192.168.0.44) at <incomplete> on wlan0
? (192.168.0.74) at <incomplete> on wlan0
? (192.168.0.185) at <incomplete> on wlan0
? (192.168.0.6) at <incomplete> on wlan0
? (192.168.0.215) at <incomplete> on wlan0
? (192.168.0.194) at <incomplete> on wlan0
? (192.168.0.49) at <incomplete> on wlan0
? (192.168.0.28) at <incomplete> on wlan0
? (192.168.0.186) at <incomplete> on wlan0
? (192.168.0.11) at <incomplete> on wlan0
? (192.168.0.216) at <incomplete> on wlan0
*/
