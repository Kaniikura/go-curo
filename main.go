package main

import (
	"fmt"
	"log"
	"net"
	"syscall"
)

func main() {
	epollFd, err := initializeEpoll()
	if err != nil {
		log.Fatalf("Failed to initialize epoll: %v", err)
	}

	netDeviceList, err := setupNetworkDevices(epollFd)
	if err != nil {
		log.Fatalf("Failed to setup network devices: %v", err)
	}

	if err := eventLoop(epollFd, netDeviceList); err != nil {
		log.Fatalf("Event loop error: %v", err)
	}
}

func initializeEpoll() (int, error) {
	return syscall.EpollCreate1(0)
}

func setupNetworkDevices(epollFd int) ([]netDevice, error) {
	var netDeviceList []netDevice
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	for _, iface := range interfaces {
		if shouldIgnoreInterface(iface) {
			continue
		}

		netDev, err := setupSingleDevice(epollFd, iface)
		if err != nil {
			return nil, fmt.Errorf("failed to setup device %s: %v", iface.Name, err)
		}

		netDeviceList = append(netDeviceList, netDev)
		fmt.Printf("Created device %s socket %d address %s\n",
			iface.Name, netDev.socket, iface.HardwareAddr.String())
	}

	return netDeviceList, nil
}

func setupSingleDevice(epollFd int, iface net.Interface) (netDevice, error) {
	sock, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
	if err != nil {
		return netDevice{}, fmt.Errorf("socket error: %v", err)
	}

	addr := syscall.SockaddrLinklayer{
		Protocol: htons(syscall.ETH_P_ALL),
		Ifindex:  iface.Index,
	}

	if err := syscall.Bind(sock, &addr); err != nil {
		return netDevice{}, fmt.Errorf("bind error: %v", err)
	}

	event := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(sock),
	}

	if err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, sock, &event); err != nil {
		return netDevice{}, fmt.Errorf("epoll ctl error: %v", err)
	}

	return createNetDevice(iface, sock, addr), nil
}

func eventLoop(epollFd int, netDeviceList []netDevice) error {
	events := make([]syscall.EpollEvent, 10)

	for {
		numEvents, err := syscall.EpollWait(epollFd, events, -1)
		if err != nil {
			return fmt.Errorf("epoll wait error: %v", err)
		}

		for i := 0; i < numEvents; i++ {
			if err := processEvent(events[i], netDeviceList); err != nil {
				return err
			}
		}
	}
}

func processEvent(event syscall.EpollEvent, netDeviceList []netDevice) error {
	for _, netDev := range netDeviceList {
		if int32(netDev.socket) == event.Fd {
			if err := netDev.receivePacket(); err != nil {
				return fmt.Errorf("failed to receive packet: %v", err)
			}
			return nil
		}
	}
	return nil
}
