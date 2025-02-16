package main

import (
	"net"
	"syscall"
)

const ETHER_TYPE_IP uint16 = 0x0800
const ETHER_TYPE_ARP uint16 = 0x0806
const ETHER_TYPE_IPV6 uint16 = 0x86dd
const ETHERNET_ADDRESS_LEN = 6

var ETHERNET_ADDRESS_BROADCAST = [6]uint8{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

type ethernetHeader struct {
	destAddr  [6]byte
	srcAddr   [6]byte
	etherType uint16
}

func (ethHdr *ethernetHeader) serialize() []byte {
	packet := make([]byte, 14)
	copy(packet[0:6], ethHdr.destAddr[:])
	copy(packet[6:12], ethHdr.srcAddr[:])
	copy(packet[12:14], uint16ToByte(ethHdr.etherType))

	return packet
}

// createPacketHandler creates a PacketHandler from network interface details.
// Takes interface, socket fd, and link-layer socket address.
// Returns configured PacketHandler struct.
func createPacketHandler(iface net.Interface, sock int, addr syscall.SockaddrLinklayer) PacketHandler {
	return PacketHandler{
		name:          iface.Name,
		macAddress:    convertMacAddress(iface.HardwareAddr),
		socketFD:      sock,
		socketAddress: addr,
	}
}

// convertMacAddress converts a net.HardwareAddr (MAC address) to a fixed-size array of 6 bytes.
// It takes a MAC address as input and returns a [6]uint8 array containing the same bytes.
// This is useful when working with raw ethernet frames where MAC addresses need to be in a fixed-size format.
func convertMacAddress(mac net.HardwareAddr) [6]uint8 {
	var macAddr [6]uint8
	copy(macAddr[:], mac)
	return macAddr
}
