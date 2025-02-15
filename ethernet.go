package main

import (
	"net"
	"syscall"
)

// type ethernetHeader struct {
// 	destAddr  [6]byte
// 	srcAddr   [6]byte
// 	etherType uint16
// }

// createNetDevice creates a netDevice from network interface details.
// Takes interface, socket fd, and link-layer socket address.
// Returns configured netDevice struct.
func createNetDevice(iface net.Interface, sock int, addr syscall.SockaddrLinklayer) netDevice {
	return netDevice{
		name:          iface.Name,
		macAddress:    convertMacAddress(iface.HardwareAddr),
		socket:        sock,
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
