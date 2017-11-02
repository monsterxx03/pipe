package main

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/monsterxx03/pipe/decoder"
	_ "github.com/monsterxx03/pipe/decoder/http"
	_ "github.com/monsterxx03/pipe/decoder/redis"
	_ "github.com/monsterxx03/pipe/decoder/text"

	"github.com/google/gopacket"
	_ "github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	localPort  = flag.String("p", "80", "Local port to capture traffic")
	traceResp  = flag.Bool("r", false, "Whether to trace response traffic")
	decodeAs   = flag.String("d", "text", "parse payload, support decoder: text, redis, http")
	deepDecode = flag.String("dd", "", "deep decode based on content type, works for http now, if -dd is provided, -d will be ignored")
	filterStr  = flag.String("f", "", "used to parse msg")
)

// eg: tcp port 80 and (host addr1 or host add2)
func buildBPFFilter(traceResp bool, localIps []string, localPort string) string {
	result := "tcp "
	if traceResp {
		result += "port " + localPort
	} else {
		// only trace incoming data
		result += "dst port " + localPort
	}
	var dstHost string
	for i, ip := range localIps {
		if traceResp {
			dstHost += " host " + ip
		} else {
			dstHost += " dst host " + ip
		}
		if i != len(localIps)-1 {
			dstHost += " or"
		}
	}
	result += " and (" + dstHost + ")"
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

func main() {
	flag.Parse()

	var wg sync.WaitGroup
	allDevs := getAlldevs()
	wg.Add(len(allDevs) + 1)

	_decodeAs := *decodeAs
	if *deepDecode != "" {
		_decodeAs = *deepDecode
	}
	d, err := decoder.GetDecoder(_decodeAs)
	if err != nil {
		panic(err)
	}
	d.SetFilter(*filterStr)

	s := NewStream(d)
	go s.To(os.Stdout)

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
				wg.Done()
				return
			}

			if err = handle.SetBPFFilter(buildBPFFilter(*traceResp, localIps, *localPort)); err != nil {
				log.Println("Failed to set BPF for:" + d.Name)
				wg.Done()
				return
			}

			packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
			for packet := range packetSource.Packets() {
				if app := packet.ApplicationLayer(); app != nil {
					_, err := s.Write(app.Payload()) // Write data to stream
					if err != nil {
						panic(err) // TODO handle error
					}
				}
			}
			wg.Done()
		}(dev)
	}
	wg.Wait()
}
