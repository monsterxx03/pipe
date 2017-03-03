package main

import (
	"flag"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"log"
	"sync"
)

var localPort = flag.String("p", "80", "Local port to capture traffic")
var to = flag.String("t", "127.0.0.1:80", "Address to send traffic")
var udp = flag.Bool("u", false, "Capture udp protocol")
var allDevices []pcap.Interface

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

func getAllIps(dev pcap.Interface) []string {
	var localIps []string
	for _, addr := range dev.Addresses {
		localIps = append(localIps, addr.IP.String())
	}
	return localIps
}

func getAlldevs() []pcap.Interface {
	if devices, err := pcap.FindAllDevs(); err != nil {
		panic(err)
	} else {
		return devices
	}
}

func main() {
	flag.Parse()
	var wg sync.WaitGroup
	allDevs := getAlldevs()
	wg.Add(len(allDevs))
	for _, dev := range allDevs {
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
				log.Println(d.Name)
				log.Println(packet)
			}
			wg.Done()
		}(dev)
	}
	wg.Wait()
}
