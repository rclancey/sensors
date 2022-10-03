package netscan

import (
	"bufio"
	"errors"
	"log"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

type MDNSRecord struct {
	Interface string
	IPVersion string
	HostName string
	ServiceName string
	Domain string
	NodeName string
	IPAddr string
	Port string
	Meta string
}

func MDNS() chan *MDNSRecord {
	ch := make(chan *MDNSRecord, 100)
	go func() {
		defer close(ch)
		cmd := exec.Command("avahi-browse", "-a", "-r", "-p", "-t")
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
			parts := strings.Split(line, ";")
			if len(parts) < 9 {
				continue
			}
			rec := &MDNSRecord{
				Interface: parts[1],
				IPVersion: parts[2],
				HostName: parts[3],
				ServiceName: parts[4],
				Domain: parts[5],
				NodeName: parts[6],
				IPAddr: parts[7],
				Port: parts[8],
			}
			if len(parts) > 9 {
				rec.Meta = parts[9]
			}
			ch <- rec
		}
		cmd.Wait()
	}()
	return ch
}

func MDNSHosts() []*HostInfo {
	hostMap := map[string]*HostInfo{}
	hosts := []*HostInfo{}
	ch := MDNS()
	for {
		rec, ok := <-ch
		if !ok {
			break
		}
		host := hostMap[rec.NodeName]
		if host == nil {
			host = NewHostInfo()
			host.Name = rec.NodeName
			host.HasMDNS = true
			hostMap[rec.NodeName] = host
			hosts = append(hosts, host)
		}
		if rec.IPVersion == "IPv6" && len(strings.Split(rec.IPAddr, ":")) > 2 {
			host.IPv6 = rec.IPAddr
		} else if rec.IPVersion == "IPv4" && len(strings.Split(rec.IPAddr, ".")) == 4 {
			host.IPv4 = rec.IPAddr
		}
		/*
		if rec.HostName != "" {
			host.Name = rec.HostName
		}
		*/
		port, err := strconv.Atoi(rec.Port)
		if err == nil {
			host.AddService(rec.ServiceName, port)
		}
	}
	for _, host := range hosts {
		if host.HasService("AirPlay Remote Video") {
			host.OS = "tvOS"
		} else if host.HasService("_sonos._tcp") {
			host.OS = "sonos"
		} else if strings.HasPrefix(host.Name, "jooki") {
			host.OS = "jooki"
		} else if host.HasService("_rdlink._tcp") {
			host.OS = "iOS"
		} else if host.HasService("Apple File Sharing") {
			host.OS = "MacOS"
		} else if host.HasService("SSH Remote Terminal") {
			host.OS = "Linux"
		}
	}
	return hosts
}

/*
pi@raspberrypi:~/src/sensors $ avahi-browse -a -r -p -t > mdns.log 
+;wlan0;IPv6;Vitalogy;Web Site;local
+;wlan0;IPv4;Vitalogy;Web Site;local
+;wlan0;IPv6;Vitalogy;Device Info;local
+;wlan0;IPv4;Vitalogy;Device Info;local
+;wlan0;IPv6;Vitalogy;iTunes Audio Access;local
+;wlan0;IPv4;Vitalogy;iTunes Audio Access;local
+;wlan0;IPv6;Vitalogy;Microsoft Windows Network;local
+;wlan0;IPv4;Vitalogy;Microsoft Windows Network;local
+;wlan0;IPv6;Sonos-949F3E898604;_sonos._tcp;local
+;wlan0;IPv4;Sonos-949F3E898604;_sonos._tcp;local
+;wlan0;IPv6;Zot;Network File System;local
+;wlan0;IPv4;Zot;Network File System;local
+;wlan0;IPv6;Zot;Apple File Sharing;local
+;wlan0;IPv6;Vitalogy;Apple File Sharing;local
+;wlan0;IPv4;Vitalogy;Apple File Sharing;local
+;wlan0;IPv4;Zot;Apple File Sharing;local
+;wlan0;IPv6;Ryan\032Clancey\226\128\153s\032Library_Music;_hscp._tcp;local
+;wlan0;IPv4;Ryan\032Clancey\226\128\153s\032Library_Music;_hscp._tcp;local
+;wlan0;IPv6;EPSON\032ET-3700\032Series\0322\032\064\032nocode;Internet Printer;local
+;wlan0;IPv4;EPSON\032ET-3700\032Series\0322\032\064\032nocode;Internet Printer;local
+;wlan0;IPv6;EPSON\032ET-3700\032Series\0322\032\064\032nocode;Secure Internet Printer;local
+;wlan0;IPv4;EPSON\032ET-3700\032Series\0322\032\064\032nocode;Secure Internet Printer;local
+;wlan0;IPv6;nocode;VNC Remote Access;local
+;wlan0;IPv6;Zot;VNC Remote Access;local
+;wlan0;IPv4;nocode;VNC Remote Access;local
+;wlan0;IPv4;Zot;VNC Remote Access;local
+;wlan0;IPv6;70-35-60-63\.1\032Living\032Room;_sleep-proxy._udp;local
+;wlan0;IPv4;70-35-60-63\.1\032Living\032Room;_sleep-proxy._udp;local
+;wlan0;IPv6;CCD2817C45EA\064Living\032Room;AirTunes Remote Audio;local
+;wlan0;IPv4;CCD2817C45EA\064Living\032Room;AirTunes Remote Audio;local
+;wlan0;IPv6;Living\032Room;AirPlay Remote Video;local
+;wlan0;IPv4;Living\032Room;AirPlay Remote Video;local
+;wlan0;IPv6;nocode;_companion-link._tcp;local
+;wlan0;IPv6;Living\032Room;_companion-link._tcp;local
+;wlan0;IPv4;nocode;_companion-link._tcp;local
+;wlan0;IPv4;Living\032Room;_companion-link._tcp;local
+;wlan0;IPv6;6105C2CB-8ABC-570C-80B1-E842013914BC;_homekit._tcp;local
+;wlan0;IPv4;6105C2CB-8ABC-570C-80B1-E842013914BC;_homekit._tcp;local
+;wlan0;IPv6;Zot;SFTP File Transfer;local
+;wlan0;IPv6;nocode;SFTP File Transfer;local
+;wlan0;IPv6;LT1565M;SFTP File Transfer;local
+;wlan0;IPv4;Zot;SFTP File Transfer;local
+;wlan0;IPv4;jooki-3BE0;SFTP File Transfer;local
+;wlan0;IPv4;nocode;SFTP File Transfer;local
+;wlan0;IPv4;LT1565M;SFTP File Transfer;local
+;wlan0;IPv6;Zot;SSH Remote Terminal;local
+;wlan0;IPv6;nocode;SSH Remote Terminal;local
+;wlan0;IPv6;LT1565M;SSH Remote Terminal;local
+;wlan0;IPv4;Zot;SSH Remote Terminal;local
+;wlan0;IPv4;jooki-3BE0;SSH Remote Terminal;local
+;wlan0;IPv4;nocode;SSH Remote Terminal;local
+;wlan0;IPv4;LT1565M;SSH Remote Terminal;local
+;wlan0;IPv6;Dad\032phone;_rdlink._tcp;local
+;wlan0;IPv4;Dad\032phone;_rdlink._tcp;local
+;wlan0;IPv4;sonos949F3E898604;_spotify-connect._tcp;local
+;wlan0;IPv4;A8EEC6003BE0;_spotify-connect._tcp;local
=;wlan0;IPv6;Vitalogy;Web Site;local;Vitalogy.local;fe80::211:32ff:fe98:2733;5000;"mac_address=00:11:32:98:27:33|00:11:32:98:27:34" "secure_admin_port=5001" "admin_port=5000" "version_build=25556" "version_minor=2" "version_major=6" "serial=1890PZN556701" "model=DS418play" "vendor=Synology"
=;wlan0;IPv4;Vitalogy;Web Site;local;Vitalogy.local;192.168.0.99;5000;"mac_address=00:11:32:98:27:33|00:11:32:98:27:34" "secure_admin_port=5001" "admin_port=5000" "version_build=25556" "version_minor=2" "version_major=6" "serial=1890PZN556701" "model=DS418play" "vendor=Synology"
=;wlan0;IPv6;Vitalogy;Device Info;local;Vitalogy.local;fe80::211:32ff:fe98:2733;0;"model=Xserve"
=;wlan0;IPv4;Vitalogy;Device Info;local;Vitalogy.local;192.168.0.99;0;"model=Xserve"
=;wlan0;IPv6;Vitalogy;iTunes Audio Access;local;Vitalogy.local;fe80::211:32ff:fe98:2733;3689;"Password=false" "Version=196610" "iTSh Version=131073" "mtd-version=0.2.4.1" "Machine Name=Vitalogy" "Machine ID=0bd1c27a" "Database ID=0bd1c27a" "txtvers=1"
=;wlan0;IPv4;Vitalogy;iTunes Audio Access;local;Vitalogy.local;192.168.0.99;3689;"Password=false" "Version=196610" "iTSh Version=131073" "mtd-version=0.2.4.1" "Machine Name=Vitalogy" "Machine ID=0bd1c27a" "Database ID=0bd1c27a" "txtvers=1"
=;wlan0;IPv6;Vitalogy;Microsoft Windows Network;local;Vitalogy.local;fe80::211:32ff:fe98:2733;445;
=;wlan0;IPv4;Vitalogy;Microsoft Windows Network;local;Vitalogy.local;192.168.0.99;445;
=;wlan0;IPv6;Sonos-949F3E898604;_sonos._tcp;local;Sonos-949F3E898604.local;192.168.0.188;1443;"protovers=1.18.9" "vers=1" "info=/api/v1/players/RINCON_949F3E89860401400/info"
=;wlan0;IPv4;Sonos-949F3E898604;_sonos._tcp;local;Sonos-949F3E898604.local;192.168.0.188;1443;"protovers=1.18.9" "vers=1" "info=/api/v1/players/RINCON_949F3E89860401400/info"
=;wlan0;IPv6;Zot;Network File System;local;Zot.local;192.168.0.64;2049;
=;wlan0;IPv4;Zot;Network File System;local;Zot.local;192.168.0.64;2049;
=;wlan0;IPv6;Zot;Apple File Sharing;local;Zot.local;192.168.0.64;548;
=;wlan0;IPv6;Vitalogy;Apple File Sharing;local;Vitalogy.local;fe80::211:32ff:fe98:2733;548;
=;wlan0;IPv4;Vitalogy;Apple File Sharing;local;Vitalogy.local;192.168.0.99;548;
=;wlan0;IPv4;Zot;Apple File Sharing;local;Zot.local;192.168.0.64;548;
=;wlan0;IPv6;Ryan\032Clancey\226\128\153s\032Library_Music;_hscp._tcp;local;nocode.local;192.168.0.151;63605;"hC=262d8ead-b2c3-4f5f-85a8-57d434ee52ee" "Machine ID=EA6BF413D6F8" "Database ID=7A040A946633C3AF" "hG=00000000-00be-0c68-8c39-9a5416c54ae9" "iCSV=65540" "OSsi=0x186A6" "Machine Name=Ryan Clancey’s Library" "DvTy=Music" "DvSv=3328" "iTSh Version=196624" "Version=196621" "MID=0xE57D4BD2D207207" "hQ=3435" "dmv=131085" "txtvers=1"
=;wlan0;IPv4;Ryan\032Clancey\226\128\153s\032Library_Music;_hscp._tcp;local;nocode.local;192.168.0.151;63605;"hC=262d8ead-b2c3-4f5f-85a8-57d434ee52ee" "Machine ID=EA6BF413D6F8" "Database ID=7A040A946633C3AF" "hG=00000000-00be-0c68-8c39-9a5416c54ae9" "iCSV=65540" "OSsi=0x186A6" "Machine Name=Ryan Clancey’s Library" "DvTy=Music" "DvSv=3328" "iTSh Version=196624" "Version=196621" "MID=0xE57D4BD2D207207" "hQ=3435" "dmv=131085" "txtvers=1"
=;wlan0;IPv6;EPSON\032ET-3700\032Series\0322\032\064\032nocode;Internet Printer;local;nocode.local;192.168.0.151;631;"printer-type=0x480900E" "printer-state=3" "Scan=T" "Color=T" "TLS=1.2" "UUID=9e854fa8-0d5d-3825-4c29-ed319404d10f" "pdl=application/octet-stream,application/pdf,application/postscript,image/jpeg,image/png,image/pwg-raster" "product=(EPSON ET-3700 Series)" "priority=0" "note=Ryan’s MacBook Pro" "adminurl=https://nocode.local.:631/printers/EPSON_ET_3700_Series_2" "ty=EPSON ET-3700 Series" "rp=printers/EPSON_ET_3700_Series_2" "qtotal=1" "txtvers=1"
=;wlan0;IPv4;EPSON\032ET-3700\032Series\0322\032\064\032nocode;Internet Printer;local;nocode.local;192.168.0.151;631;"printer-type=0x480900E" "printer-state=3" "Scan=T" "Color=T" "TLS=1.2" "UUID=9e854fa8-0d5d-3825-4c29-ed319404d10f" "pdl=application/octet-stream,application/pdf,application/postscript,image/jpeg,image/png,image/pwg-raster" "product=(EPSON ET-3700 Series)" "priority=0" "note=Ryan’s MacBook Pro" "adminurl=https://nocode.local.:631/printers/EPSON_ET_3700_Series_2" "ty=EPSON ET-3700 Series" "rp=printers/EPSON_ET_3700_Series_2" "qtotal=1" "txtvers=1"
=;wlan0;IPv6;EPSON\032ET-3700\032Series\0322\032\064\032nocode;Secure Internet Printer;local;nocode.local;192.168.0.151;631;"printer-type=0x480900E" "printer-state=3" "Scan=T" "Color=T" "TLS=1.2" "UUID=9e854fa8-0d5d-3825-4c29-ed319404d10f" "pdl=application/octet-stream,application/pdf,application/postscript,image/jpeg,image/png,image/pwg-raster" "product=(EPSON ET-3700 Series)" "priority=0" "note=Ryan’s MacBook Pro" "adminurl=https://nocode.local.:631/printers/EPSON_ET_3700_Series_2" "ty=EPSON ET-3700 Series" "rp=printers/EPSON_ET_3700_Series_2" "qtotal=1" "txtvers=1"
=;wlan0;IPv4;EPSON\032ET-3700\032Series\0322\032\064\032nocode;Secure Internet Printer;local;nocode.local;192.168.0.151;631;"printer-type=0x480900E" "printer-state=3" "Scan=T" "Color=T" "TLS=1.2" "UUID=9e854fa8-0d5d-3825-4c29-ed319404d10f" "pdl=application/octet-stream,application/pdf,application/postscript,image/jpeg,image/png,image/pwg-raster" "product=(EPSON ET-3700 Series)" "priority=0" "note=Ryan’s MacBook Pro" "adminurl=https://nocode.local.:631/printers/EPSON_ET_3700_Series_2" "ty=EPSON ET-3700 Series" "rp=printers/EPSON_ET_3700_Series_2" "qtotal=1" "txtvers=1"
=;wlan0;IPv6;nocode;VNC Remote Access;local;nocode.local;192.168.0.151;5900;
=;wlan0;IPv6;Zot;VNC Remote Access;local;Zot.local;192.168.0.64;5900;
=;wlan0;IPv4;nocode;VNC Remote Access;local;nocode.local;192.168.0.151;5900;
=;wlan0;IPv4;Zot;VNC Remote Access;local;Zot.local;192.168.0.64;5900;
=;wlan0;IPv6;70-35-60-63\.1\032Living\032Room;_sleep-proxy._udp;local;Living-Room.local;192.168.0.139;52993;
=;wlan0;IPv4;70-35-60-63\.1\032Living\032Room;_sleep-proxy._udp;local;Living-Room.local;192.168.0.139;52993;
=;wlan0;IPv6;CCD2817C45EA\064Living\032Room;AirTunes Remote Audio;local;Living-Room.local;192.168.0.139;7000;"vv=2" "ov=16.0" "vs=635.87.3" "vn=65537" "tp=UDP" "pk=777be3e75b63ddd7c0f11faf96ed9bae4a5fcb0e97c7561c09b967334d72802f" "am=AppleTV6,2" "md=0,1,2" "sf=0x244" "ft=0x4A7FDFD5,0xBC157FDE" "et=0,3,5" "da=true" "cn=0,1,2,3"
=;wlan0;IPv4;CCD2817C45EA\064Living\032Room;AirTunes Remote Audio;local;Living-Room.local;192.168.0.139;7000;"vv=2" "ov=16.0" "vs=635.87.3" "vn=65537" "tp=UDP" "pk=777be3e75b63ddd7c0f11faf96ed9bae4a5fcb0e97c7561c09b967334d72802f" "am=AppleTV6,2" "md=0,1,2" "sf=0x244" "ft=0x4A7FDFD5,0xBC157FDE" "et=0,3,5" "da=true" "cn=0,1,2,3"
=;wlan0;IPv6;Living\032Room;AirPlay Remote Video;local;Living-Room.local;192.168.0.139;7000;"vv=2" "osvers=16.0" "srcvers=635.87.3" "pk=777be3e75b63ddd7c0f11faf96ed9bae4a5fcb0e97c7561c09b967334d72802f" "psi=C14C9E01-6FF3-4FB9-A006-99AEF6F00669" "pi=8abd2ce2-44c2-46da-b87b-07659d02f44b" "protovers=1.1" "model=AppleTV6,2" "gcgl=1" "igl=1" "gid=E5318DEA-27AC-4E97-92DC-5D293EEE9831" "flags=0x244" "features=0x4A7FDFD5,0xBC157FDE" "fex=1d9/St5/FbwooQ" "deviceid=CC:D2:81:7C:45:EA" "btaddr=0A:ED:FE:80:3A:A7" "acl=0"
=;wlan0;IPv4;Living\032Room;AirPlay Remote Video;local;Living-Room.local;192.168.0.139;7000;"vv=2" "osvers=16.0" "srcvers=635.87.3" "pk=777be3e75b63ddd7c0f11faf96ed9bae4a5fcb0e97c7561c09b967334d72802f" "psi=C14C9E01-6FF3-4FB9-A006-99AEF6F00669" "pi=8abd2ce2-44c2-46da-b87b-07659d02f44b" "protovers=1.1" "model=AppleTV6,2" "gcgl=1" "igl=1" "gid=E5318DEA-27AC-4E97-92DC-5D293EEE9831" "flags=0x244" "features=0x4A7FDFD5,0xBC157FDE" "fex=1d9/St5/FbwooQ" "deviceid=CC:D2:81:7C:45:EA" "btaddr=0A:ED:FE:80:3A:A7" "acl=0"
=;wlan0;IPv6;nocode;_companion-link._tcp;local;nocode.local;192.168.0.151;51292;"rpBA=18:22:C7:EB:6C:18" "rpHI=f4c065c551be" "rpAD=cbd52fd13cf6" "rpHA=bb9a08a8f2f2" "rpVr=195.2" "rpFl=0x20000" "rpHN=740eb8b16992"
=;wlan0;IPv6;Living\032Room;_companion-link._tcp;local;Living-Room.local;192.168.0.139;49153;"rpBA=FA:DF:34:D4:AA:6A" "rpAD=001c144c8549" "rpMRtID=C14C9E01-6FF3-4FB9-A006-99AEF6F00669" "rpVr=400.51" "rpMd=AppleTV6,2" "rpFl=0xB6782" "rpHN=96b33d89c896" "rpMac=2"
=;wlan0;IPv4;nocode;_companion-link._tcp;local;nocode.local;192.168.0.151;51292;"rpBA=18:22:C7:EB:6C:18" "rpHI=f4c065c551be" "rpAD=cbd52fd13cf6" "rpHA=bb9a08a8f2f2" "rpVr=195.2" "rpFl=0x20000" "rpHN=740eb8b16992"
=;wlan0;IPv4;Living\032Room;_companion-link._tcp;local;Living-Room.local;192.168.0.139;49153;"rpBA=FA:DF:34:D4:AA:6A" "rpAD=001c144c8549" "rpMRtID=C14C9E01-6FF3-4FB9-A006-99AEF6F00669" "rpVr=400.51" "rpMd=AppleTV6,2" "rpFl=0xB6782" "rpHN=96b33d89c896" "rpMac=2"
=;wlan0;IPv6;6105C2CB-8ABC-570C-80B1-E842013914BC;_homekit._tcp;local;Living-Room.local;192.168.0.139;59805;"si=40ABA0E6-3A18-4280-BBB1-D4FF7E06CF93"
=;wlan0;IPv4;6105C2CB-8ABC-570C-80B1-E842013914BC;_homekit._tcp;local;Living-Room.local;192.168.0.139;59805;"si=40ABA0E6-3A18-4280-BBB1-D4FF7E06CF93"
=;wlan0;IPv6;Zot;SFTP File Transfer;local;Zot.local;192.168.0.64;22;
=;wlan0;IPv6;nocode;SFTP File Transfer;local;nocode.local;192.168.0.151;22;
=;wlan0;IPv6;LT1565M;SFTP File Transfer;local;LT1565M.local;192.168.0.123;22;
=;wlan0;IPv4;Zot;SFTP File Transfer;local;Zot.local;192.168.0.64;22;
=;wlan0;IPv4;jooki-3BE0;SFTP File Transfer;local;jooki-3BE0.local;192.168.0.210;22;
=;wlan0;IPv4;nocode;SFTP File Transfer;local;nocode.local;192.168.0.151;22;
=;wlan0;IPv4;LT1565M;SFTP File Transfer;local;LT1565M.local;192.168.0.123;22;
=;wlan0;IPv6;Zot;SSH Remote Terminal;local;Zot.local;192.168.0.64;22;
=;wlan0;IPv6;nocode;SSH Remote Terminal;local;nocode.local;192.168.0.151;22;
=;wlan0;IPv6;LT1565M;SSH Remote Terminal;local;LT1565M.local;192.168.0.123;22;
=;wlan0;IPv4;Zot;SSH Remote Terminal;local;Zot.local;192.168.0.64;22;
=;wlan0;IPv4;jooki-3BE0;SSH Remote Terminal;local;jooki-3BE0.local;192.168.0.210;22;
=;wlan0;IPv4;nocode;SSH Remote Terminal;local;nocode.local;192.168.0.151;22;
=;wlan0;IPv4;LT1565M;SSH Remote Terminal;local;LT1565M.local;192.168.0.123;22;
=;wlan0;IPv6;Dad\032phone;_rdlink._tcp;local;Dad-phone.local;192.168.0.243;49159;"rpAD=945bf09140f3" "rpVr=360.4" "rpBA=69:BA:0A:C7:C5:75"
=;wlan0;IPv4;Dad\032phone;_rdlink._tcp;local;Dad-phone.local;192.168.0.243;49159;"rpAD=945bf09140f3" "rpVr=360.4" "rpBA=69:BA:0A:C7:C5:75"
=;wlan0;IPv4;sonos949F3E898604;_spotify-connect._tcp;local;sonos949F3E898604.local;192.168.0.188;1400;"CPath=/spotifyzc" "VERSION=1.0"
=;wlan0;IPv4;A8EEC6003BE0;_spotify-connect._tcp;local;A8EEC6003BE0.local;192.168.0.210;8080;"Stack=SP" "VERSION=1.0" "CPath=/zc/0"
*/
