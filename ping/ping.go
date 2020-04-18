package ping

import (
	"log"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// Ping struct hold user configuration and runtime data
// that used for ping'ing library
type Ping struct {

	// Wait interval seconds between sending each packet.
	// The default is to wait for one second between each packet normally
	// Only super-user may set interval to values less 0.2 seconds.
	Interval int

	// IPAddress are extracted from Target which may come from a domain names
	// or directly from user input, the value can be IPv4/IPv6
	IPAddress string

	// Listen IPv4 or IPv6 format that used for listening incoming message
	Listen string

	// Message is a ICMP payload
	Message []byte

	// Network contain udp4 & udp6
	Network string

	// ProtocolNumber specify which protocol (IP-Based) are used
	// as ping target.
	// https://en.wikipedia.org/wiki/List_of_IP_protocol_numbers
	ProtocolNumber int

	// RRT used to store every ping RRT time
	RRT []int64

	// Sequence store current instance ping sequence
	Sequence int

	// Success store total succeed ICMP Request
	Success int

	// Failed store total failed ICMP Request
	Failed int

	// Target are plain user-input that only allow domain names and
	// IP Address (IPv4/IPv6)
	Target string

	// Time to wait for a response, in seconds, any request that
	// longer than Timeout will be closed automatically
	Timeout int

	// Set the IP Time to Live or Hop Limit for IPv6
	// based on [1]; the default recommended value are 64
	// 1: https://en.wikipedia.org/wiki/Time_to_live
	TTL int
}

// Ping later
func (p *Ping) Ping() PingResult {

	// increase sequence
	p.Sequence++

	// hold current ping result
	result := PingResult{}

	// create ICMP packet listener that used for giving back
	// reply to our machine
	c, err := icmp.ListenPacket(p.Network, p.Listen)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	if p.ProtocolNumber == 1 {
		c.IPv4PacketConn().SetTTL(p.TTL)
	} else {
		c.IPv6PacketConn().SetHopLimit(p.TTL)
	}

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: p.Sequence,
			Data: p.Message,
		},
	}

	if p.ProtocolNumber == 58 {
		wm.Type = ipv6.ICMPTypeEchoRequest
	}

	wb, err := wm.Marshal(nil)
	if err != nil {
		log.Println(err)
		p.Failed++
		return result
	}

	// save when we call c.WriteTo
	start := time.Now()
	if _, err := c.WriteTo(wb, &net.UDPAddr{IP: net.ParseIP(p.IPAddress), Zone: "en0"}); err != nil {
		log.Println(err)
		p.Failed++
		return result
	}

	rb := make([]byte, 1500)
	n, _, err := c.ReadFrom(rb)
	if err != nil {
		log.Println(err)
		p.Failed++
		return result
	}
	// save when we got the __call back__ from target
	elapsed := time.Since(start)

	rm, err := icmp.ParseMessage(p.ProtocolNumber, rb[:n])
	if err != nil {
		log.Println(err)
		p.Failed++
		return result
	}

	usedTTL := 0
	retPayloadSize := 0

	if p.ProtocolNumber == 1 {
		retPayloadSize = rm.Body.Len(p.ProtocolNumber)
		usedTTL, _ = c.IPv4PacketConn().TTL()
	} else {
		// if target is a IPv6
		retPayloadSize = rm.Body.Len(58)
		usedTTL, _ = c.IPv6PacketConn().HopLimit()
	}

	result.UsedTTL = usedTTL
	result.RTT = elapsed
	result.PayloadSize = retPayloadSize
	result.Message = rb
	result.Success = true

	// missunderstand
	if strings.Contains(string(rb[:n]), "time exceeded") {
		// fmt.Println("must uwu")
		// fmt.Println(string(rb))
	}

	p.RRT = append(p.RRT, elapsed.Microseconds())

	p.Success++

	return result
}
