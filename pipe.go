package main

import (
	"flag"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
	"sync"
)

var localPort = flag.String("p", "80", "Local port to capture traffic")
var to = flag.String("t", "stdout", "Address to send traffic, stdout, 127.0.0.1:8080 ....")
var decode = flag.String("d", "ascii", "parse payload, support decoder: ascii, redis, mysql")
var udp = flag.Bool("u", false, "Capture udp protocol")
var allDevices []pcap.Interface

// eg: tcp dst port 80 and (dst host addr1 or dst host add2)
// only monitor incoming traffic
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
	log.Println(result)
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

func getAlldevs() []pcap.Interface {
	if devices, err := pcap.FindAllDevs(); err != nil {
		panic(err)
	} else {
		return devices
	}
}

func getAllIps(dev pcap.Interface) []string {
	var localIps []string
	for _, addr := range dev.Addresses {
		localIps = append(localIps, addr.IP.String())
	}
	return localIps
}

func connect(udp bool, addr string) net.Conn {
	var protocol string
	if udp {
		protocol = "udp"
	} else {
		protocol = "tcp"
	}
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		log.Fatal("Failed to connect to remote:" + addr)
	}
	return conn
}

func main() {
	flag.Parse()
	var wg sync.WaitGroup
	allDevs := getAlldevs()
	wg.Add(len(allDevs))

	// Replace with Output plugin
	conn := connect(*udp, *to)
	defer conn.Close()

	for _, dev := range allDevs {
		// use one goroutine for every device
		go func(d pcap.Interface) {
			handle, err := pcap.OpenLive(d.Name, 65536, true, pcap.BlockForever)
			if err != nil {
				log.Println("fail to listen:" + d.Name)
				wg.Done()
				return
			}
			defer handle.Close()

			var localIps []string
			if localIps = getAllIps(d); len(localIps) == 0 {
				log.Println("No ip found for:" + d.Name)
				wg.Done()
				return
			}

			if err = handle.SetBPFFilter(buildBPFFilter(*udp, localIps, *localPort)); err != nil {
				log.Println("Failed to set BPF for:" + d.Name)
				wg.Done()
				return
			}

			packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
			for packet := range packetSource.Packets() {
				if aL := packet.ApplicationLayer(); aL != nil {
					conn.Write(aL.Payload())
				}
			}
			wg.Done()
		}(dev)
	}
	wg.Wait()
}
