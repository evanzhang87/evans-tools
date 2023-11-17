package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"io"
	"log"
)

const (
	headerLength      = 16
	DubboVersionOrder = 0
	PathOrder         = 1
	VersionOrder      = 2
	MethodOrder       = 3
)

var (
	pcapFile string
)

func init() {
	flag.StringVar(&pcapFile, "f", "test.pcap", "pcap file path")
}

type simpleReader struct {
	reader *bufio.Reader
	header []byte
	offset int
	body   []byte
}

func main() {
	flag.Parse()
	handle, err := pcap.OpenOffline(pcapFile)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()
	packerSource := gopacket.NewPacketSource(handle, handle.LinkType())
	n := 0
	for packet := range packerSource.Packets() {
		if packet.ApplicationLayer() != nil {
			data := packet.ApplicationLayer().Payload()
			bufReader := &simpleReader{
				reader: bufio.NewReader(bytes.NewReader(data)),
				header: make([]byte, headerLength),
				offset: 0,
			}
			bodyLen, skip := bufReader.readHeader()
			if bodyLen > 0 {
				if skip {
					bufReader.reader.Discard(bodyLen)
					continue
				}
				if bufReader.body == nil {
					bufReader.body = make([]byte, bodyLen)
				} else {
					if bodyLen > len(bufReader.body) {
						bufReader.body = nil
						bufReader.body = make([]byte, bodyLen)
					} else {
						bufReader.body = bufReader.body[:bodyLen]
					}
				}
				metaMap := bufReader.readBody()
				if metaMap == nil {
					continue
				}
				dubboVersion := metaMap[DubboVersionOrder]
				path := metaMap[PathOrder]
				version := metaMap[VersionOrder]
				method := metaMap[MethodOrder]
				n += 1
				fmt.Printf("dubbo version: %s, path: %s, version: %s, method: %s \n", dubboVersion, path, version, method)
			} else {
				fmt.Println("can't get info from header, skip")
				continue
			}
		}
	}
	fmt.Println("package count:", n)
}

func (r *simpleReader) readHeader() (int, bool) {
	n, err := r.reader.Read(r.header)
	if n < headerLength {
		_, err = r.reader.Read(r.header[n:])
	}
	if err == io.EOF {
		fmt.Println("Read Header EOF")
	} else if err != nil {
		fmt.Printf("Read Header err: %v \n", err)
	} else {
		bodyLen := int(r.header[12])<<24 + int(r.header[13])<<16 + int(r.header[14])<<8 + int(r.header[15])
		if r.header[0] == 0xda && r.header[1] == 0xbb {
			if r.header[2]&0x80 == 0 {
				return r.reader.Buffered(), true
			}
			if r.header[2]&0x20 != 0 {
				return bodyLen, true
			}
			return bodyLen, false
		} else {
			return r.reader.Buffered(), true
		}
	}
	return 0, true
}

func (r *simpleReader) readBody() map[int]string {
	_, err := r.reader.Read(r.body)
	if err == io.EOF {
		fmt.Println("Read Body EOF")
		return nil
	} else if err != nil {
		fmt.Println("Read Body err", err)
		return nil
	} else {
		r.offset = 0
		metaMap := make(map[int]string, 4)
		for i := 0; i < 4; i++ {
			flag := r.body[r.offset]
			r.offset += 1
			if flag <= 0x1f {
				metaMap[i] = string(r.body[r.offset : r.offset+int(flag)])
				r.offset += int(flag)
			} else if flag >= 0x30 && flag <= 0x33 {
				buflen := int(flag-0x30)*256 + int(r.body[r.offset])
				r.offset += 1
				metaMap[i] = string(r.body[r.offset : r.offset+buflen])
				r.offset += buflen
			} else if flag == 'N' {
				metaMap[i] = "Null"
				r.offset += 1
			} else if flag == 'T' {
				metaMap[i] = "True"
				r.offset += 1
			} else if flag == 'F' {
				metaMap[i] = "False"
				r.offset += 1
			} else {
				metaMap[i] = "unknown"
				r.offset += 1
			}
		}
		return metaMap
	}
}
