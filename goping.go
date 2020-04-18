// Cloudflare Internship Task
// Reference :
// https://godoc.org/golang.org/x/net/icmp
// https://en.wikipedia.org/wiki/List_of_IP_protocol_numbers
// https://en.wikipedia.org/wiki/Time_to_live
// https://linux.die.net/man/8/ping

package main

import (
	"flag"
	"fmt"
	"goping/helper"
	"goping/ping"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var targetIP string

func main() {

	var interval int
	flag.IntVar(&interval, "i", 1, "request interval delay in second")

	var timeout int
	flag.IntVar(&timeout, "t", 10, "request timeout in second")

	var ttl int
	flag.IntVar(&ttl, "ttl", 64, "set IP Time To Live")

	quiet := flag.Bool("q", false, "Quiet output. Nothing is displayed except summary when finished.")
	html := flag.Bool("html", false, "Save HTML Output (ping statistic)")

	flag.Parse()

	plainTarget := ""

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) == 0 {
		fmt.Println("usage: ./goping google.com")
		fmt.Println("usage: ./goping -i 2 -t 3 -html google.com")
		flag.PrintDefaults()
		os.Exit(1)
	}

	for _, arg := range argsWithoutProg {
		if arg == "help" {
			flag.PrintDefaults()
			break
		}
		// put plain arg as plainTarget
		plainTarget = arg
		// check is arg a domain name
		if helper.IsDomainName(arg) {
			addr, err := net.LookupIP(arg)
			if err != nil {
				panic(err)
			} else {
				targetIP = addr[0].String()
			}
			break
		}
		// check is arg a valid IP Address
		if res := net.ParseIP(arg); res != nil {
			targetIP = arg
			break
		}
	}

	if targetIP == "" {
		panic("Invalid Target (IP Address/Domain name)")
	}

	// create new ping instance
	p := ping.Ping{
		Interval:       interval,
		IPAddress:      targetIP,
		Listen:         "0.0.0.0",
		Network:        "udp4",
		ProtocolNumber: 1,
		Target:         plainTarget,
		Timeout:        timeout,
		TTL:            ttl,
		Message:        []byte("echo requests"),
	}

	// set IPv6 Network & ProtocolNumber if target IP Address
	// was IPv6
	if helper.IsIPv6(targetIP) {
		p.Network = "udp6"
		p.ProtocolNumber = 58
		p.Listen = "::"
	}

	var start time.Time
	// listen for exit signal and print ping statistic
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		elapsed := time.Since(start)
		// print header
		fmt.Printf("\n--- %v ping statistics ---\n", p.Target)
		// calculate loss
		loss := 100 - (p.Success / p.Sequence * 100)
		fmt.Printf("%v packets transmitted, %v received, %v%% packet loss, time %s\n", p.Sequence, p.Success, loss, elapsed)
		// calculate RTT Information
		min, max := helper.MinMax(p.RRT)
		// calculate average RTT
		var total int64 = 0
		for _, value := range p.RRT {
			total += value
		}
		avg := total / int64(len(p.RRT))
		fmt.Printf("rtt min/avg/max = %v/%v/%v ms\n", min/1000, avg/1000, max/1000)
		// generate HTML Output if html option was called
		if *html {
			filename, err := helper.GenHTML(p.Target, p.RRT, p.Sequence, start)
			if err != nil {
				fmt.Println("Failed to save HTML Output file, reason : %v", err.Error())
			} else {
				fmt.Println("successfully saved html files in : " + filename)
			}
		}
		os.Exit(0)
	}()

	start = time.Now()
	for {

		done := make(chan bool)
		res := make(chan ping.PingResult, 1)

		go func() {
			res <- p.Ping()
			done <- true
		}()

	G:
		select {
		case <-done:
			if *quiet {
				break G
			}
			data := <-res
			// incrase/decrease RRT percetage
			percentChange := float64(0)
			if len(p.RRT) > 1 {
				// current request RTT in milliseconds
				reqMS := data.RTT.Microseconds()
				before := p.RRT[len(p.RRT)-2]
				percentChange = helper.PercentageChange(before, reqMS)
			}
			diff := fmt.Sprintf("%f", percentChange)
			if percentChange > 0 {
				diff = "+" + diff
			}
			fmt.Printf("%v bytes from %v: icmp_seq=%v ttl=%v time=%s \t [%v%%]\n", data.PayloadSize, p.IPAddress, p.Sequence, data.UsedTTL, data.RTT, diff)
		case <-time.After(time.Duration(timeout) * time.Second):
			if *quiet {
				break G
			}
			fmt.Printf("From %v icmp_seq=%v Destination Host Unreachable\n", p.IPAddress, p.Sequence)
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}
