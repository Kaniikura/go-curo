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

	pHandlerList, err := setupPacketHandlers(epollFd)
	if err != nil {
		log.Fatalf("Failed to setup network devices: %v", err)
	}

	if err := eventLoop(epollFd, pHandlerList); err != nil {
		log.Fatalf("Event loop error: %v", err)
	}
}

func initializeEpoll() (int, error) {
	return syscall.EpollCreate1(0)
}

// setupPacketHandlers initializes network devices for packet handling.
// Creates raw sockets for each network interface and sets up epoll monitoring.
//
// Parameters:
//   - epollFd: File descriptor for epoll
//
// Returns:
//   - []PacketHandler: Slice of initialized devices
//   - error: Error if setup fails
//
// Filters out interfaces that should be ignored and configures remaining ones.
func setupPacketHandlers(epollFd int) ([]PacketHandler, error) {
	var pHandlerList []PacketHandler
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	for _, iface := range interfaces {
		if shouldIgnoreInterface(iface) {
			continue
		}

		pHandler, err := setupSingleHandler(epollFd, iface)
		if err != nil {
			return nil, fmt.Errorf("failed to setup device %s: %v", iface.Name, err)
		}

		pHandlerList = append(pHandlerList, pHandler)
		fmt.Printf("Created device %s socket %d address %s\n",
			iface.Name, pHandler.socketFD, iface.HardwareAddr.String())
	}

	return pHandlerList, nil
}

// setupSingleHandler sets up a network device for packet capture.
// Creates raw socket, binds it to interface, adds to epoll monitoring.
//
// Parameters:
//   - epollFd: Epoll file descriptor
//   - iface: Network interface
//
// Returns:
//   - PacketHandler: Configured device
//   - error: Setup error
func setupSingleHandler(epollFd int, iface net.Interface) (PacketHandler, error) {
	sock, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
	if err != nil {
		return PacketHandler{}, fmt.Errorf("socket error: %v", err)
	}

	addr := syscall.SockaddrLinklayer{
		Protocol: htons(syscall.ETH_P_ALL),
		Ifindex:  iface.Index,
	}

	if err := syscall.Bind(sock, &addr); err != nil {
		return PacketHandler{}, fmt.Errorf("bind error: %v", err)
	}

	event := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(sock),
	}

	if err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, sock, &event); err != nil {
		return PacketHandler{}, fmt.Errorf("epoll ctl error: %v", err)
	}

	return createPacketHandler(iface, sock, addr), nil
}

// eventLoop is the main event processing loop that handles network device events.
// It continuously waits for events from the epoll instance and processes them.
//
// Parameters:
//   - epollFd: File descriptor for the epoll instance
//   - pHandlerList: Slice of network devices to monitor
//
// Returns:
//   - error: Returns an error if epoll wait fails or event processing fails
//
// The function blocks indefinitely waiting for events and only returns on error.
func eventLoop(epollFd int, pHandlerList []PacketHandler) error {
	events := make([]syscall.EpollEvent, 10)

	for {
		numEvents, err := syscall.EpollWait(epollFd, events, -1)
		if err != nil {
			return fmt.Errorf("epoll wait error: %v", err)
		}

		for i := 0; i < numEvents; i++ {
			if err := processEvent(events[i], pHandlerList); err != nil {
				return err
			}
		}
	}
}

// processEvent handles an epoll event by finding the corresponding network device
// from the provided list and processing any incoming packets. It iterates through
// the network device list to find a matching file descriptor, then calls receiveFrame
// on the matching device.
//
// Parameters:
//   - event: The epoll event containing the file descriptor and event flags
//   - pHandlerList: Slice of network devices to check against the event
//
// Returns:
//   - error: nil on success, or an error if packet reception fails
func processEvent(event syscall.EpollEvent, pHandlerList []PacketHandler) error {
	for _, pHandler := range pHandlerList {
		if int32(pHandler.socketFD) == event.Fd {
			if err := pHandler.receiveFrame(); err != nil {
				return fmt.Errorf("failed to receive packet: %v", err)
			}
			return nil
		}
	}
	return nil
}
