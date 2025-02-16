package main

import (
	"encoding/binary"
	"fmt"
)

func bytesToUint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func bytesToUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func uint16ToByte(i uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	return b
}

func uint32ToByte(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return b
}

func uint32ToOctets(i uint32) [4]uint8 {
	return [4]uint8{uint8(i >> 24), uint8(i >> 16), uint8(i >> 8), uint8(i)}
}

func formatMACAddress(mac [6]uint8) string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}
