package main

import (
	"fmt"
	"net"
	"syscall"
)

var ignoredInterfaces = []string{"lo", "bond0", "dummy0", "tunl0", "sit0"}

// netDevice represents a network device.
type netDevice struct {
	name          string
	macAddress    [6]uint8
	socket        int
	socketAddress syscall.SockaddrLinklayer
	// ethernetHdr   ethernetHeader
	// ipDev ipDevice
}

// shouldIgnoreInterface checks if a network interface should be ignored
// based on the ignoredInterfaces list.
// Returns true if interface should be ignored, false otherwise.
func shouldIgnoreInterface(iface net.Interface) bool {
	for i := range ignoredInterfaces {
		if iface.Name == ignoredInterfaces[i] {
			return true
		}
	}
	return false
}

// htons converts a uint16 to network byte order.
func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}

const maxEthernetFrameSize = 1500

// receivePacket receives a single Ethernet frame from the network device.
// It allocates a buffer of maximum Ethernet frame size and uses syscall.Recvfrom
// to read incoming data from the device's socket.
//
// The function handles the following cases:
// - If bytes read is -1, returns nil (no data available)
// - If there's an error reading from socket, returns error with device name
// - On successful read, prints received bytes for debugging
//
// Returns an error if packet reception fails, nil otherwise.
func (netDev *netDevice) receivePacket() error {
	frameBuffer := make([]byte, maxEthernetFrameSize)
	bytesRead, _, err := syscall.Recvfrom(netDev.socket, frameBuffer, 0)

	if err != nil {
		if bytesRead == -1 {
			return nil
		}
		return fmt.Errorf("failed to receive packet on interface %s: %w", netDev.name, err)
	}

	fmt.Printf("Received %d bytes from %s: %x\n", bytesRead, netDev.name, frameBuffer[:bytesRead])
	return nil
}
