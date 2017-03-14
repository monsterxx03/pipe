package main

import (
	"flag"
	"log"
	"net"
	"runtime/debug"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var (
	localPort  = flag.String("p", "80", "Local port to capture traffic")
	to         = flag.String("t", "", "Address to send traffic, stdout, 127.0.0.1:8080 ....")
	traceResp  = flag.Bool("r", false, "Whether to trace response traffic")
	decodeAs   = flag.String("d", "", "parse payload, support decoder: ascii, redis, mysql")
	udp        = flag.Bool("u", false, "Capture udp protocol")
	allDevices []pcap.Interface
	mode       string = "decode" // decode or mirror
)

// eg: tcp port 80 and (host addr1 or host add2)
func buildBPFFilter(traceResp bool, udp bool, localIps []string, localPort string) string {
	result := ""
	if udp {
		result += "udp "
	} else {
		result += "tcp "
	}
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

func init() {
	flag.Parse()
	if *decodeAs != "" {
		mode = "decode"
	} else {
		if *to == "" {
			log.Fatal("Must provide -t or -d")
		}
		log.Println("mirror:" + *to)
		mode = "mirror"
	}
}

func handlePacket(conn net.Conn, packet gopacket.Packet, localPort string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered in handlePacket:", r)
			debug.PrintStack()
		}
	}()
	if mode == "decode" {
		if net := packet.TransportLayer(); net != nil {
			// decode transport layer to get port info
			var direction string
			srcPort, _ := net.TransportFlow().Endpoints()
			if srcPort.String() == localPort {
				direction = "resp: >>>"
			} else {
				direction = "req: <<<"
			}

			if aL := packet.ApplicationLayer(); aL != nil {
				if data, err := decode(*decodeAs, aL.Payload()); err != nil {
					log.Println("Failed to decode:", err)
				} else {
					log.Println(direction + data)
				}
			}
		}
	} else {
		// mirror to remote
		if aL := packet.ApplicationLayer(); aL != nil {
			conn.Write(aL.Payload())
		}
	}
}

func main() {
	var wg sync.WaitGroup
	allDevs := getAlldevs()
	wg.Add(len(allDevs))

	var conn net.Conn
	if mode == "mirror" {
		conn = connect(*udp, *to)
		defer conn.Close()
	}

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

			if err = handle.SetBPFFilter(buildBPFFilter(*traceResp, *udp, localIps, *localPort)); err != nil {
				log.Println("Failed to set BPF for:" + d.Name)
				wg.Done()
				return
			}

			packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
			for packet := range packetSource.Packets() {
				handlePacket(conn, packet, *localPort)
			}
			wg.Done()
		}(dev)
	}
	wg.Wait()
}
