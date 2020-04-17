package ping

import "time"

// PingResult will hold data
type PingResult struct {

	// UsedTTL
	UsedTTL int

	// Message that returned from target
	Message []byte

	// RTT are Round-Trip Delay Time,
	RTT time.Duration

	// Success set as true when an ICMP Request succeed
	Success bool

	// PayloadSize hold returned payload data in bytes
	PayloadSize int
}
