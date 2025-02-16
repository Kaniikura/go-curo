package main

import (
	"fmt"
	"net"
	"syscall"
)

const maxEthernetFrameSize = 1500

type PacketHandler struct {
	name          string
	macAddress    [6]uint8
	socketFD      int
	socketAddress syscall.SockaddrLinklayer
	ethernetHdr   ethernetHeader
	ipv4Config    IPv4Interface
}

// sendFrame sends a single Ethernet frame to the network device.
func (p *PacketHandler) sendFrame(frame []byte) {
	_, err := syscall.Write(p.socketFD, frame)
	if err != nil {
		fmt.Printf("Failed to send frame: %v\n", err)
	}
}

// receiveFrame reads a single Ethernet frame from the network interface associated with the PacketHandler.
// It receives the frame into a buffer, parses the Ethernet header, checks if the packet is destined for the device,
// and then handles the packet based on its EtherType. Currently, it only handles ARP packets.
//
// Returns:
//   - error: An error if there was a problem receiving the frame, or nil if the frame was received successfully.
//     If syscall.Recvfrom returns -1, it is considered non-fatal and nil is returned.
func (pHandler *PacketHandler) receiveFrame() error {
	frameBuffer := make([]byte, maxEthernetFrameSize)
	bytesRead, _, err := syscall.Recvfrom(pHandler.socketFD, frameBuffer, 0)

	if err != nil {
		if bytesRead == -1 {
			return nil
		}
		return fmt.Errorf("failed to receive packet on interface %s: %w", pHandler.name, err)
	}

	packet := frameBuffer[:bytesRead]

	fmt.Printf("Received %d bytes from %s: %x\n", bytesRead, pHandler.name, frameBuffer[:bytesRead])

	// Parse Ethernet header
	pHandler.ethernetHdr.destAddr = convertMacAddress(packet[0:6])
	pHandler.ethernetHdr.srcAddr = convertMacAddress(packet[6:12])
	pHandler.ethernetHdr.etherType = bytesToUint16(packet[12:14])

	// Check if packet is destined for this device
	if pHandler.macAddress != pHandler.ethernetHdr.destAddr && pHandler.ethernetHdr.destAddr != ETHERNET_ADDRESS_BROADCAST {
		return nil
	}

	// Handle packet based on EtherType
	switch pHandler.ethernetHdr.etherType {
	case ETHER_TYPE_ARP:
		handleARPPacket(*pHandler, packet[14:])
		// case ETHER_TYPE_IP:
		// 	handleIPPacket(*pHandler, packet[14:])
	}

	return nil
}

var ignoredInterfaces = []string{"lo", "bond0", "dummy0", "tunl0", "sit0"}

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
