// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

/*
 * Sample DHCPv4 client to test on the local interface
 */

import (
	"flag"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/toujourser/coredhcp/logger"
)

var log = logger.GetLogger("main")

var (
	serverIP = flag.String("server", "255.255.255.255", "DHCP server IP address")
	localIP  = flag.String("local", "0.0.0.0", "Local IP address to bind to")
)

func main() {
	flag.Parse()

	var macString string
	if len(flag.Args()) > 0 {
		macString = flag.Arg(0)
	} else {
		macString = "fa:16:3e:ac:e6:e5"
	}

	mac, err := net.ParseMAC(macString)
	if err != nil {
		log.Fatal(err)
	}

	// Create UDP connection
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.ParseIP(*localIP),
		Port: 68,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Enable broadcast
	if err := conn.SetWriteBuffer(1024 * 1024); err != nil {
		log.Printf("Warning: failed to set write buffer: %v", err)
	}

	serverAddr := &net.UDPAddr{
		IP:   net.ParseIP(*serverIP),
		Port: 67,
	}

	log.Printf("Client listening on %s:68", *localIP)
	log.Printf("Server address: %s:67", *serverIP)

	// Create DHCP Discover message with broadcast flag
	discover, err := dhcpv4.NewDiscovery(mac,
		dhcpv4.WithBroadcast(true),
		dhcpv4.WithRequestedOptions(
			dhcpv4.OptionSubnetMask,
			dhcpv4.OptionRouter,
			dhcpv4.OptionDomainNameServer,
		))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Sending DISCOVER: %s", discover.Summary())

	// Send Discover
	if _, err := conn.WriteToUDP(discover.ToBytes(), serverAddr); err != nil {
		log.Fatal(err)
	}

	// Receive Offer
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buffer := make([]byte, 1500)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		log.Fatal(err)
	}

	offer, err := dhcpv4.FromBytes(buffer[:n])
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Received OFFER: %s", offer.Summary())

	// Create DHCP Request message with broadcast flag
	request, err := dhcpv4.NewRequestFromOffer(offer, dhcpv4.WithBroadcast(true))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Sending REQUEST: %s", request.Summary())

	// Send Request
	if _, err := conn.WriteToUDP(request.ToBytes(), serverAddr); err != nil {
		log.Fatal(err)
	}

	// Receive Ack
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, _, err = conn.ReadFromUDP(buffer)
	if err != nil {
		log.Fatal(err)
	}

	ack, err := dhcpv4.FromBytes(buffer[:n])
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Received ACK: %s", ack.Summary())
	log.Printf("DHCP exchange completed successfully!")
}
