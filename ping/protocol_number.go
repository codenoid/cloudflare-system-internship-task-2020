package ping

const (
	// UNSPECIFIED logs nothing
	UNSPECIFIED ProtocolNumber = iota // 0 :
	// TRACE logs everything
	TRACE // 1
	// INFO logs Info, Warnings and Errors
	INFO // 2
	// WARNING logs Warning and Errors
	WARNING // 3
	// ERROR just logs Errors
	ERROR // 4
)

// ProtocolNumber for ICMP must only have 2 value
// 1 for ipv4 and 58 for ipv6
type ProtocolNumber int
