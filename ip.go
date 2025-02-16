package main

import (
	"fmt"
)

type IPv4Interface struct {
	ipAddress uint32
	// networkMask   uint32
	// broadcastAddr uint32
	// natDevice   natDevice
}

func formatIPv4Address(ipAddress uint32) string {
	octets := uint32ToOctets(ipAddress)
	return fmt.Sprintf("%d.%d.%d.%d", octets[0], octets[1], octets[2], octets[3])
}
