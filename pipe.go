package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"runtime/debug"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

var (
	localPort    = flag.String("p", "80", "Local port to capture traffic")
	traceResp    = flag.Bool("r", false, "Whether to trace response traffic")
	decodeAs     = flag.String("d", "", "parse payload, support decoder: ascii, redis, http")
	writeToFile  = flag.String("w", "", "Write payload to file")
	filterStr    = flag.String("f", "", "used to parse msg")
	silence      = flag.Bool("s", false, "silence output")
	enableStream = flag.Bool("stream", false, "track packet in stream mode(via tcpassembly)")
	allDevices   []pcap.Interface
	localFile    *os.File = nil
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
	if *writeToFile != "" {
		localFile, err = os.OpenFile(*writeToFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}
	}
}

type tcpStreamFactory struct{}

// httpStream will handle the actual decoding of http requests.
type stream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
}

func (s *tcpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	stream := &stream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
	}
	go stream.run() // Important... we must guarantee that data from the reader stream is read.

	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return &stream.r
}

func (s *stream) run() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered in stream run:", r)
			debug.PrintStack()
		}
	}()
	reader := bufio.NewReader(&s.r)
	decoder, err := getDecoder(*decodeAs, *filterStr)
	if err != nil {
		panic(err)
	}
	decoder.SetReader(reader)
	for {
		data, err := decoder.Decode()
		if err == io.EOF {
			return
		} else if err == SkipError {
			continue
		} else if err != nil {
			log.Println(err)
		} else {
			log.Println(data)
		}
	}
}

func handlePacket(packet gopacket.Packet) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered in handlePacket:", r)
			debug.PrintStack()
		}
	}()
	if app := packet.ApplicationLayer(); app != nil {
		reader := bufio.NewReader(bytes.NewReader(app.Payload()))
		decoder, err := getDecoder(*decodeAs, *filterStr)
		if err != nil {
			panic(err)
		}
		decoder.SetReader(reader)
		data, err := decoder.Decode()
		if err == io.EOF || err == SkipError {
			return
		} else if err != nil {
			log.Println(err)
		} else {
			log.Println(data)
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
			// make tcpassembler
			var assembler *tcpassembly.Assembler
			if *enableStream {
				streamFactory := &tcpStreamFactory{}
				streamPool := tcpassembly.NewStreamPool(streamFactory)
				assembler = tcpassembly.NewAssembler(streamPool)
			}

			for packet := range packetSource.Packets() {
				if *enableStream {
					tcp := packet.TransportLayer().(*layers.TCP)
					assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)
				} else {
					// handle packet one by one
					handlePacket(packet)
				}
			}
			wg.Done()
		}(dev)
	}
	wg.Wait()
}
