package main

import (
	"fmt"
)

const ARP_OPERATION_CODE_REQUEST = 1
const ARP_OPERATION_CODE_REPLY = 2
const ARP_HTYPE_ETHERNET uint16 = 0001

// ArpTable is a slice of ARP entries.
var ArpTable []ArpEntry

// ArpEntry represents an entry in the ARP table.
type ArpEntry struct {
	ipAddress  uint32
	macAddress [6]uint8
}

// ARPPacket represents an ARP packet.
type ARPPacket struct {
	hType     uint16   // Hardware Type (e.g., Ethernet)
	pType     uint16   // Protocol Type (e.g., IPv4)
	hLen      uint8    // Hardware Address Length
	pLen      uint8    // Protocol Address Length
	operation uint16   // Operation Code (request/reply)
	srcMAC    [6]uint8 // Source MAC Address
	srcIP     uint32   // Source IP Address
	dstMAC    [6]uint8 // Destination MAC Address
	dstIP     uint32   // Destination IP Address
}

func (arp ARPPacket) serialize() []byte {
	packet := make([]byte, 28)
	copy(packet[0:2], uint16ToByte(arp.hType))
	copy(packet[2:4], uint16ToByte(arp.pType))
	packet[4] = arp.hLen
	packet[5] = arp.pLen
	copy(packet[6:8], uint16ToByte(arp.operation))
	copy(packet[8:14], arp.srcMAC[:])
	copy(packet[14:18], uint32ToByte(arp.srcIP))
	copy(packet[18:24], arp.dstMAC[:])
	copy(packet[24:28], uint32ToByte(arp.dstIP))

	return packet
}

func handleARPPacket(packetHandler PacketHandler, packet []byte) {
	if len(packet) < 28 {
		fmt.Println("ARP packet too short")
		return
	}

	// Parse ARP packet
	arpMessage := ARPPacket{
		hType:     bytesToUint16(packet[0:2]),
		pType:     bytesToUint16(packet[2:4]),
		hLen:      packet[4],
		pLen:      packet[5],
		operation: bytesToUint16(packet[6:8]),
		srcMAC:    convertMacAddress(packet[8:14]),
		srcIP:     bytesToUint32(packet[14:18]),
		dstMAC:    convertMacAddress(packet[18:24]),
		dstIP:     bytesToUint32(packet[24:28]),
	}

	switch arpMessage.pType {
	case ETHER_TYPE_IP:
		if arpMessage.hLen != ETHERNET_ADDRESS_LEN {
			fmt.Println("ARP hardware address length mismatch")
			return
		}

		if arpMessage.pLen != 4 {
			fmt.Println("ARP protocol address length mismatch")
			return
		}

		if arpMessage.operation == ARP_OPERATION_CODE_REQUEST {
			fmt.Printf("ARP Request from %v for %v\n", formatIPv4Address(arpMessage.srcIP), formatIPv4Address(arpMessage.dstIP))
			receiveArpRequest(packetHandler, arpMessage)
		} else if arpMessage.operation == ARP_OPERATION_CODE_REPLY {
			fmt.Printf("ARP Reply Packet is %+v\n", arpMessage)
			receiveArpReply(packetHandler, arpMessage)
		}
	}
}

func receiveArpRequest(pHandler PacketHandler, arpPacket ARPPacket) {
	// Check if ARP request is for this device
	if arpPacket.dstIP == pHandler.ipv4Config.ipAddress {
		fmt.Printf("Sending ARP reply to %v\n", formatIPv4Address(arpPacket.srcIP))
		arpPacket := buildArpReply(pHandler, arpPacket)
		sendArpReply(pHandler, arpPacket)
	}
}

func receiveArpReply(pHandler PacketHandler, arpPacket ARPPacket) {
	// Check if ARP reply is for this device
	if pHandler.ipv4Config.ipAddress != 00000000 {
		fmt.Printf("Adding ARP entry by arp reply (%s -> %s)\n", formatIPv4Address(arpPacket.srcIP), formatMACAddress(arpPacket.srcMAC))
		// Add ARP entry
		addArpEntry(arpPacket.srcIP, arpPacket.srcMAC)
	}
}

func addArpEntry(ip uint32, mac [6]uint8) {
	// if table is empty, create a new entry
	if len(ArpTable) == 0 {
		createArpEntry(ip, mac)
		return
	}

	// check if entry already exists
	for _, entry := range ArpTable {
		// if ip address already exists, update the mac address
		if entry.ipAddress == ip {
			entry.macAddress = mac
			return
		}
		// if mac address already exists, update the ip address
		if entry.macAddress == mac {
			entry.ipAddress = ip
			return
		}
	}

	// if entry does not exist, create a new entry
	createArpEntry(ip, mac)
}

func createArpEntry(ip uint32, mac [6]uint8) {
	arpEntry := ArpEntry{
		ipAddress:  ip,
		macAddress: mac,
	}

	ArpTable = append(ArpTable, arpEntry)
}

func buildArpReply(pHandler PacketHandler, arpMessage ARPPacket) ARPPacket {
	arpReply := ARPPacket{
		hType:     ARP_HTYPE_ETHERNET,
		pType:     ETHER_TYPE_IP,
		hLen:      ETHERNET_ADDRESS_LEN,
		pLen:      4,
		operation: ARP_OPERATION_CODE_REPLY,
		srcMAC:    pHandler.macAddress,
		srcIP:     pHandler.ipv4Config.ipAddress,
		dstMAC:    arpMessage.srcMAC,
		dstIP:     arpMessage.srcIP,
	}

	return arpReply
}

func sendArpReply(pHandler PacketHandler, arpPacket ARPPacket) {
	ethernetHeader := ethernetHeader{
		destAddr:  arpPacket.dstMAC,
		srcAddr:   pHandler.macAddress,
		etherType: ETHER_TYPE_ARP,
	}

	ethernetFrame := append(ethernetHeader.serialize(), arpPacket.serialize()...)
	pHandler.sendFrame(ethernetFrame)
}
