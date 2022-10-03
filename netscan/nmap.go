package netscan

import (
	//"encoding/json"
	"encoding/xml"
	"log"
	//"os"
	"os/exec"
)

type NMAPRun struct {
	XMLName xml.Name `xml:"nmaprun"`
	Hosts []NMAPHost `xml:"host"`
}

type NMAPHost struct {
	Addrs []NMAPAddress `xml:"address"`
	Ports []NMAPPort `xml:"ports>porrt"`
	OSMatches []NMAPOSMatch `xml:"os>osmatch"`
}

func (host *NMAPHost) GetOS() string {
	for _, osmatch := range host.OSMatches {
		for _, osclass := range osmatch.OSClasses {
			return osclass.Family
		}
	}
	return ""
}

type NMAPAddress struct {
	XMLName xml.Name `xml:"address"`
	Addr string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

type NMAPPort struct {
	XMLName xml.Name `xml:"port"`
	Proto string `xml:"protocol,attr"`
	Port int `xml:"portid,attr"`
	Name string `xml:"service>name,attr"`
	State string `xml:"state>state,attr"`
}

type NMAPOSMatch struct {
	XMLName xml.Name `xml:"osmatch"`
	Name string `xml:"name,attr"`
	OSClasses []NMAPOSClass `xml:"osclass"`
	Accuracy int `xml:"accuracy,attr"`
}

type NMAPOSClass struct {
	Type string `xml:"type,attr"`
	Vendor string `xml:"vendor,attr"`
	Family string `xml:"osfamily,attr"`
	Version string `xml:"osgen,attr"`
	Accuracy int `xml:"accuracy,attr"`
}

func ParseNMAP(xmldata []byte) (*NMAPRun, error) {
	run := &NMAPRun{}
	err := xml.Unmarshal(xmldata, run)
	if err != nil {
		log.Println(string(xmldata))
		log.Fatalln("xml error:", err)
		return nil, err
	}
	return run, nil
}

func RawNMAP(ip string) (*NMAPRun, error) {
	log.Println("nmap", ip)
	cmd := exec.Command("sudo", "nmap", "-sS", "-O", "-oX", "-", "-p", "ssh,http,https,iphone-sync", ip)
	data, err := cmd.Output()
	if err != nil {
		xerr, ok := err.(*exec.ExitError)
		if ok {
			log.Println("error", string(xerr.Stderr))
		}
		return nil, err
	}
	run, err := ParseNMAP(data)
	if err != nil {
		return nil, err
	}
	//enc := json.NewEncoder(os.Stderr)
	//enc.Encode(run)
	return run, nil
}

func NMAP(ip string) ([]*HostInfo, error) {
	run, err := RawNMAP(ip)
	if err != nil {
		return nil, err
	}
	hosts := []*HostInfo{}
	for _, nhost := range run.Hosts {
		host := NewHostInfo()
		host.HasNMAP = true
		hosts = append(hosts, host)
		for _, addr := range nhost.Addrs {
			switch addr.AddrType {
			case "mac":
				host.MAC = addr.Addr
			case "ipv4":
				host.IPv4 = addr.Addr
			case "ipv6":
				host.IPv6 = addr.Addr
			}
		}
		for _, port := range nhost.Ports {
			if port.State == "open" {
				host.AddService(port.Name, port.Port)
			}
		}
		if host.HasService("iphone-sync") {
			host.OS = "iOS"
		} else {
			host.OS = nhost.GetOS()
		}
	}
	return hosts, nil
}

/*
pi@raspberrypi:~/src/sensors $ sudo nmap -sS -O -oX - -p ssh,http,https,iphone-sync 192.168.0.243
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE nmaprun>
<?xml-stylesheet href="file:///usr/bin/../share/nmap/nmap.xsl" type="text/xsl"?>
<!-- Nmap 7.80 scan initiated Sun Sep 25 15:20:19 2022 as: nmap -sS -O -oX - -p ssh,http,https,iphone-sync 192.168.0.243 -->
<nmaprun scanner="nmap" args="nmap -sS -O -oX - -p ssh,http,https,iphone-sync 192.168.0.243" start="1664144419" startstr="Sun Sep 25 15:20:19 2022" version="7.80" xmloutputversion="1.04">
<scaninfo type="syn" protocol="tcp" numservices="5" services="22,80,443,8008,62078"/>
<verbose level="0"/>
<debugging level="0"/>
<host starttime="1664144419" endtime="1664144442"><status state="up" reason="arp-response" reason_ttl="0"/>
<address addr="192.168.0.243" addrtype="ipv4"/>
<address addr="C6:65:E9:67:B0:2C" addrtype="mac"/>
<hostnames>
</hostnames>
<ports><port protocol="tcp" portid="22"><state state="closed" reason="reset" reason_ttl="64"/><service name="ssh" method="table" conf="3"/></port>
<port protocol="tcp" portid="80"><state state="closed" reason="reset" reason_ttl="64"/><service name="http" method="table" conf="3"/></port>
<port protocol="tcp" portid="443"><state state="closed" reason="reset" reason_ttl="64"/><service name="https" method="table" conf="3"/></port>
<port protocol="tcp" portid="8008"><state state="closed" reason="reset" reason_ttl="64"/><service name="http" method="table" conf="3"/></port>
<port protocol="tcp" portid="62078"><state state="open" reason="syn-ack" reason_ttl="64"/><service name="iphone-sync" method="table" conf="3"/></port>
</ports>
<os><portused state="open" proto="tcp" portid="62078"/>
<portused state="closed" proto="tcp" portid="22"/>
<portused state="closed" proto="udp" portid="44594"/>
<osmatch name="Apple Mac OS X 10.7.0 (Lion) - 10.12 (Sierra) or iOS 4.1 - 9.3.3 (Darwin 10.0.0 - 16.4.0)" accuracy="91" line="6488">
<osclass type="general purpose" vendor="Apple" osfamily="Mac OS X" osgen="10.7.X" accuracy="91"><cpe>cpe:/o:apple:mac_os_x:10.7</cpe></osclass>
<osclass type="general purpose" vendor="Apple" osfamily="OS X" osgen="10.8.X" accuracy="91"><cpe>cpe:/o:apple:mac_os_x:10.8</cpe></osclass>
<osclass type="general purpose" vendor="Apple" osfamily="OS X" osgen="10.9.X" accuracy="91"><cpe>cpe:/o:apple:mac_os_x:10.9</cpe></osclass>
<osclass type="general purpose" vendor="Apple" osfamily="OS X" osgen="10.10.X" accuracy="91"><cpe>cpe:/o:apple:mac_os_x:10.10</cpe></osclass>
<osclass type="general purpose" vendor="Apple" osfamily="OS X" osgen="10.11.X" accuracy="91"><cpe>cpe:/o:apple:mac_os_x:10.11</cpe></osclass>
<osclass type="general purpose" vendor="Apple" osfamily="macOS" osgen="10.12.X" accuracy="91"><cpe>cpe:/o:apple:mac_os_x:10.12</cpe></osclass>
<osclass type="media device" vendor="Apple" osfamily="iOS" osgen="4.X" accuracy="91"><cpe>cpe:/o:apple:iphone_os:4</cpe><cpe>cpe:/a:apple:apple_tv:4</cpe></osclass>
<osclass type="phone" vendor="Apple" osfamily="iOS" osgen="4.X" accuracy="91"><cpe>cpe:/o:apple:iphone_os:4</cpe></osclass>
<osclass type="phone" vendor="Apple" osfamily="iOS" osgen="5.X" accuracy="91"><cpe>cpe:/o:apple:iphone_os:5</cpe></osclass>
<osclass type="phone" vendor="Apple" osfamily="iOS" osgen="6.X" accuracy="91"><cpe>cpe:/o:apple:iphone_os:6</cpe></osclass>
<osclass type="phone" vendor="Apple" osfamily="iOS" osgen="7.X" accuracy="91"><cpe>cpe:/o:apple:iphone_os:7</cpe></osclass>
<osclass type="phone" vendor="Apple" osfamily="iOS" osgen="8.X" accuracy="91"><cpe>cpe:/o:apple:iphone_os:8</cpe></osclass>
<osclass type="phone" vendor="Apple" osfamily="iOS" osgen="9.X" accuracy="91"><cpe>cpe:/o:apple:iphone_os:9</cpe></osclass>
</osmatch>
<osmatch name="Apple iOS 9.0 (Darwin 15.0.0)" accuracy="90" line="4403">
<osclass type="phone" vendor="Apple" osfamily="iOS" osgen="9.X" accuracy="90"><cpe>cpe:/o:apple:iphone_os:9.0</cpe></osclass>
</osmatch>
<osmatch name="Apple Mac OS X Server 10.5 (Leopard) pre-release build 9A284" accuracy="90" line="6053">
<osclass type="general purpose" vendor="Apple" osfamily="Mac OS X" osgen="10.5.X" accuracy="90"><cpe>cpe:/o:apple:mac_os_x_server:10.5</cpe></osclass>
</osmatch>
<osmatch name="Apple Mac OS X 10.4.11 (Tiger) (Darwin 8.11.0, PowerPC)" accuracy="89" line="5469">
<osclass type="general purpose" vendor="Apple" osfamily="Mac OS X" osgen="10.4.X" accuracy="89"><cpe>cpe:/o:apple:mac_os_x:10.4.11</cpe></osclass>
</osmatch>
<osmatch name="Apple TV 5.2.1 or 5.3" accuracy="89" line="3360">
<osclass type="media device" vendor="Apple" osfamily="Apple TV" osgen="5.X" accuracy="89"><cpe>cpe:/a:apple:apple_tv:5.2.1</cpe><cpe>cpe:/a:apple:apple_tv:5.3</cpe></osclass>
</osmatch>
<osmatch name="Apple OS X 10.11 (El Capitan) - 10.12 (Sierra) or iOS 10.1 - 10.2 (Darwin 15.4.0 - 16.6.0)" accuracy="88" line="7509">
<osclass type="general purpose" vendor="Apple" osfamily="OS X" osgen="10.11.X" accuracy="88"><cpe>cpe:/o:apple:mac_os_x:10.11</cpe></osclass>
<osclass type="general purpose" vendor="Apple" osfamily="macOS" osgen="10.12.X" accuracy="88"><cpe>cpe:/o:apple:mac_os_x:10.12</cpe></osclass>
<osclass type="phone" vendor="Apple" osfamily="iOS" osgen="10.X" accuracy="88"><cpe>cpe:/o:apple:iphone_os:10</cpe></osclass>
</osmatch>
<osmatch name="DragonFly BSD 4.6-RELEASE" accuracy="87" line="23385">
<osclass type="general purpose" vendor="DragonFly BSD" osfamily="DragonFly BSD" osgen="4.X" accuracy="87"><cpe>cpe:/o:dragonflybsd:dragonfly_bsd:4.6</cpe></osclass>
</osmatch>
<osmatch name="Epson Stylus Pro 400 printer" accuracy="87" line="24598">
<osclass type="printer" vendor="Epson" osfamily="embedded" accuracy="87"><cpe>cpe:/h:epson:stylus_pro_400</cpe></osclass>
</osmatch>
<osmatch name="Apple iOS 11.0" accuracy="87" line="3607">
<osclass type="phone" vendor="Apple" osfamily="iOS" osgen="11.X" accuracy="87"><cpe>cpe:/o:apple:iphone_os:11.0</cpe></osclass>
</osmatch>
<osmatch name="Apple iPad tablet computer (iOS 4.3.2)" accuracy="87" line="3663">
<osclass type="media device" vendor="Apple" osfamily="iOS" osgen="4.X" accuracy="87"><cpe>cpe:/o:apple:iphone_os:4.3.2</cpe></osclass>
</osmatch>
<osfingerprint fingerprint="OS:SCAN(V=7.80%E=4%D=9/25%OT=62078%CT=22%CU=44594%PV=Y%DS=1%DC=D%G=Y%M=C665&#xa;OS:E9%TM=6330D43B%P=arm-unknown-linux-gnueabihf)SEQ(SP=107%GCD=1%ISR=109%TI&#xa;OS:=Z%CI=RD%TS=22)OPS(O1=M5B4NW5NNT11SLL%O2=M5B4NW5NNT11SLL%O3=M5B4NW5NNT11&#xa;OS:%O4=M5B4NW5NNT11SLL%O5=M5B4NW5NNT11SLL%O6=M5B4NNT11SLL)WIN(W1=B50%W2=AD8&#xa;OS:%W3=4E8%W4=FFFF%W5=418%W6=1FA)ECN(R=Y%DF=Y%T=40%W=0%O=M5B4NW5SLL%CC=N%Q=&#xa;OS:)T1(R=Y%DF=Y%T=40%S=O%A=S+%F=AS%RD=0%Q=)T2(R=N)T3(R=N)T4(R=Y%DF=Y%T=40%W&#xa;OS:=0%S=A%A=Z%F=R%O=%RD=0%Q=)T5(R=Y%DF=N%T=40%W=0%S=Z%A=S+%F=AR%O=%RD=0%Q=)&#xa;OS:T6(R=Y%DF=Y%T=40%W=0%S=A%A=Z%F=R%O=%RD=0%Q=)T7(R=Y%DF=N%T=40%W=0%S=Z%A=S&#xa;OS:%F=AR%O=%RD=0%Q=)U1(R=Y%DF=N%T=40%IPL=38%UN=0%RIPL=G%RID=G%RIPCK=G%RUCK=&#xa;OS:0%RUD=G)IE(R=Y%DFI=S%T=40%CD=S)&#xa;"/>
</os>
<uptime seconds="2" lastboot="Sun Sep 25 15:20:41 2022"/>
<distance value="1"/>
<tcpsequence index="260" difficulty="Good luck!" values="BEFBD09B,60DDABA,61A7E76F,73C5F76D,193EE10F,CFBD9A"/>
<ipidsequence class="All zeros" values="0,0,0,0,0,0"/>
<tcptssequence class="other" values="C2B1A16C,D23A4543,D1102CA,40F093E1,893B0DBB,5D44BD10"/>
<times srtt="36203" rttvar="15277" to="100000"/>
</host>
<runstats><finished time="1664144443" timestr="Sun Sep 25 15:20:43 2022" elapsed="23.90" summary="Nmap done at Sun Sep 25 15:20:43 2022; 1 IP address (1 host up) scanned in 23.90 seconds" exit="success"/><hosts up="1" down="0" total="1"/>
</runstats>
</nmaprun>
*/
