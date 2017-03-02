package main

import (
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var iface = flag.String("i", "eth0", "Interface to read packets from")
var localPort = flag.String("f", "80", "Local port to capture traffic")
var to = flag.String("t", "127.0.0.1:80", "Address to send traffic")
var udp = flag.Bool("u", false, "Capture udp protocol")

// tcp dst port 80 and (dst host addr1 or dst host add2)
func buildBPFFilter(udp bool, localIps []string, localPort string) string {
	result := ""
	if udp {
		result += "udp "
	} else {
		result += "tcp "
	}
	result += "dst port " + localPort
	var dstHost string
	for i, ip := range localIps {
		dstHost += " dst host " + ip
		if i != len(localIps)-1 {
			dstHost += " or "
		}
	}
	result += " and (" + dstHost + ")"
	fmt.Println(result)
	return result
}

func getDev(devName string) pcap.Interface {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		panic(err)
	}
	for _, d := range devices {
		if d.Name == devName {
			return d
		}
	}
	panic("Failed to find interface: " + devName)
}

func main() {
	flag.Parse()
	var localIps []string
	for _, add := range getDev(*iface).Addresses {
		localIps = append(localIps, add.IP.String())
	}
	if len(localIps) == 0 {
		panic("no ip found")
	}
	handle, err := pcap.OpenLive(*iface, 1600, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}
	defer handle.Close()

	err = handle.SetBPFFilter(buildBPFFilter(*udp, localIps, *localPort))
	if err != nil {
		panic(err)
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		fmt.Println(packet)
	}
}
