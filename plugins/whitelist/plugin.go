// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// Package whitelist implements a MAC address whitelist plugin for CoreDHCP.
// Only MAC addresses listed in the configuration will be allowed to receive IP assignments.
// If no whitelist is configured, there are no restrictions.
package whitelist

import (
	"strings"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/toujourser/coredhcp/handler"
	"github.com/toujourser/coredhcp/logger"
	"github.com/toujourser/coredhcp/plugins"
)

var log = logger.GetLogger("plugins/whitelist")

// Plugin wraps the information necessary to register a plugin.
var Plugin = plugins.Plugin{
	Name:   "whitelist",
	Setup6: setup6,
	Setup4: setup4,
}

// whitelistedMACs holds the list of MAC addresses that are allowed to receive IP assignments
var whitelistedMACs map[string]bool

// Handler6 handles DHCPv6 packets for the whitelist plugin
func Handler6(req, resp dhcpv6.DHCPv6) (dhcpv6.DHCPv6, bool) {
	// If no whitelist is configured, allow all
	if whitelistedMACs == nil || len(whitelistedMACs) == 0 {
		return resp, false
	}

	mac, err := dhcpv6.ExtractMAC(req)
	if err != nil {
		log.Warningf("Could not extract MAC address from DHCPv6 packet: %v", err)
		// Drop the request if we can't extract MAC address
		return nil, true
	}

	macStr := strings.ToLower(mac.String())
	if !whitelistedMACs[macStr] {
		log.Infof("MAC address %s is not in whitelist, dropping request", macStr)
		// Drop the request if MAC is not in whitelist
		return nil, true
	}

	log.Debugf("MAC address %s is in whitelist, allowing request", macStr)
	return resp, false
}

// Handler4 handles DHCPv4 packets for the whitelist plugin
func Handler4(req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
	// If no whitelist is configured, allow all
	if whitelistedMACs == nil || len(whitelistedMACs) == 0 {
		return resp, false
	}

	macStr := strings.ToLower(req.ClientHWAddr.String())
	log.Printf("[DHCP-whitelist] MAC address %s is in whitelist, allowing request", macStr)
	if !whitelistedMACs[macStr] {
		log.Infof("MAC address %s is not in whitelist, dropping request", macStr)
		// Drop the request if MAC is not in whitelist
		return nil, true
	}

	log.Debugf("MAC address %s is in whitelist, allowing request", macStr)
	return resp, false
}

func setup6(args ...string) (handler.Handler6, error) {
	log.Printf("[DHCP-whitelist] loading `whitelist` plugin for DHCPv6")
	log.Printf("[DHCP-whitelist] Args length=%d, content=%+v", len(args), args)

	// Parse MAC addresses from args
	whitelistedMACs = make(map[string]bool)
	for i, arg := range args {
		log.Printf("Processing arg[%d]: '%s'", i, arg)
		mac := strings.ToLower(strings.TrimSpace(arg))
		// Validate MAC address format (simple validation)
		if len(mac) > 0 {
			log.Printf("[DHCP-whitelist] Adding MAC to whitelist: '%s'", mac)
			whitelistedMACs[mac] = true
		} else {
			log.Printf("[DHCP-whitelist] Skipping empty MAC address at arg[%d]", i)
		}
	}

	log.Printf("[DHCP-whitelist] Final whitelistedMACs content: %+v, %d", whitelistedMACs, len(whitelistedMACs))
	return Handler6, nil
}

func setup4(args ...string) (handler.Handler4, error) {
	log.Printf("[DHCP-whitelist] loading `whitelist` plugin for DHCPv6")
	log.Printf("[DHCP-whitelist] Args length=%d, content=%+v", len(args), args)

	// For DHCPv4, we use the same whitelist as DHCPv6
	// If this is the first setup call, parse the MAC addresses
	if whitelistedMACs == nil {
		log.Printf("[DHCP-whitelist] Initializing whitelistedMACs map")
		whitelistedMACs = make(map[string]bool)
		for i, arg := range args {
			log.Printf("[DHCP-whitelist] Processing arg[%d]: '%s'", i, arg)
			mac := strings.ToLower(strings.TrimSpace(arg))
			// Validate MAC address format (simple validation)
			if len(mac) > 0 {
				log.Printf("[DHCP-whitelist] Adding MAC to whitelist: '%s'", mac)
				whitelistedMACs[mac] = true
			} else {
				log.Printf("[DHCP-whitelist] Skipping empty MAC address at arg[%d]", i)
			}
		}
	} else {
		log.Printf("[DHCP-whitelist] whitelistedMACs already initialized with %d entries", len(whitelistedMACs))
	}

	log.Printf("[DHCP-whitelist] Final whitelistedMACs content: %+v, %d", whitelistedMACs, len(whitelistedMACs))
	return Handler4, nil
}
