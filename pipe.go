package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"runtime/debug"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var (
	localPort   = flag.String("p", "80", "Local port to capture traffic")
	to          = flag.String("t", "", "Address to send traffic, stdout, 127.0.0.1:8080 ....")
	traceResp   = flag.Bool("r", false, "Whether to trace response traffic")
	decodeAs    = flag.String("d", "", "parse payload, support decoder: ascii, redis, mysql")
	writeToFile = flag.String("w", "", "Write payload to file")
	filterStr   = flag.String("f", "", "used to parse msg")
	silence     = flag.Bool("s", false, "silence output")
	allDevices  []pcap.Interface
	localFile   *os.File = nil
	decoder     Decoder  = nil
	conn        net.Conn = nil
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

func InitCli() {
	flag.Parse()
	var err error
	if *decodeAs != "" {
		decoder, err = getDecoder(*decodeAs, *filterStr)
		if err != nil {
			panic(err)
		}
	} else if *to == "" {
		panic("Need -d or -t must be provided at least one")
	}
	if *writeToFile != "" {
		localFile, err = os.OpenFile(*writeToFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}
	}
}

func handlePacket(packet gopacket.Packet, localPort string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered in handlePacket:", r)
			debug.PrintStack()
		}
	}()
	if decoder != nil {
		// if decoder and filter exist, need to decode transport payload to do filter, else just mirror to remote
		if net := packet.TransportLayer(); net != nil {
			// decode transport layer to get port info
			var direction string
			srcPort, _ := net.TransportFlow().Endpoints()
			if srcPort.String() == localPort {
				direction = "resp:\n"
			} else {
				direction = "req:\n"
			}

			if aL := packet.ApplicationLayer(); aL != nil {
				payload := aL.Payload()
				if data, err := decoder.Decode(payload); err != nil {
					if err != SkipError && err != io.EOF {
						log.Println("Failed to decode:", err)
					}
				} else {
					if !*silence {
						log.Printf("%v %q\n", direction, data)
					}
					if *to != "" {
						writeToRemote(payload)
					}
					if localFile != nil {
						writePayload([]byte(data))
					}
				}
			}
		}
	} else {
		// mirror to remote
		if aL := packet.ApplicationLayer(); aL != nil {
			payload := aL.Payload()
			if localFile != nil {
				writePayload(payload)
			}
			writeToRemote(payload)
		}
	}
}

func writePayload(data []byte) {
	_, err := localFile.Write(data)
	if err != nil {
		panic(err)
	}
	_, err = localFile.Write([]byte{'\n'})
	if err != nil {
		panic(err)
	}
}

func writeToRemote(data []byte) {
	forceNew := false
	if len(data) > 0 {
		for i := 0; i < 3; i++ {
			conn, err := getConnection(forceNew)
			if err != nil {
				log.Println(err)
				continue
			}
			_, err = conn.Write(data)
			if err != nil {
				log.Println(err)
				forceNew = true
				continue
			}
			break
		}
	}
}

func getConnection(force bool) (net.Conn, error) {
	var err error
	if conn == nil || force {
		conn, err = net.Dial("tcp", *to)
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}

func main() {
	InitCli()
	var wg sync.WaitGroup
	allDevs := getAlldevs()
	wg.Add(len(allDevs))

	if localFile != nil {
		defer localFile.Close()
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

			if err = handle.SetBPFFilter(buildBPFFilter(*traceResp, localIps, *localPort)); err != nil {
				log.Println("Failed to set BPF for:" + d.Name)
				wg.Done()
				return
			}

			packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
			for packet := range packetSource.Packets() {
				handlePacket(packet, *localPort)
			}
			wg.Done()
		}(dev)
	}
	wg.Wait()
}
