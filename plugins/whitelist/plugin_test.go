// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package whitelist

import (
	"net"
	"testing"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/stretchr/testify/assert"
)

func TestWhitelistSetup(t *testing.T) {
	// Test DHCPv4 setup
	handler4, err := setup4("00:11:22:33:44:55", "aa:bb:cc:dd:ee:ff")
	assert.NoError(t, err)
	assert.NotNil(t, handler4)

	// Test DHCPv6 setup
	handler6, err := setup6("00:11:22:33:44:55", "aa:bb:cc:dd:ee:ff")
	assert.NoError(t, err)
	assert.NotNil(t, handler6)

	// Check that the whitelist was populated correctly
	assert.Equal(t, 2, len(whitelistedMACs))
	assert.True(t, whitelistedMACs["00:11:22:33:44:55"])
	assert.True(t, whitelistedMACs["aa:bb:cc:dd:ee:ff"])
}

func TestWhitelistHandler4(t *testing.T) {
	// Setup whitelist with specific MAC addresses
	_, err := setup4("00:11:22:33:44:55", "aa:bb:cc:dd:ee:ff")
	assert.NoError(t, err)

	// Create a DHCPv4 request with a whitelisted MAC
	mac, err := net.ParseMAC("00:11:22:33:44:55")
	assert.NoError(t, err)
	req, err := dhcpv4.NewDiscovery(mac)
	assert.NoError(t, err)

	// Process the request - should be allowed
	resp, err := dhcpv4.NewReplyFromRequest(req)
	assert.NoError(t, err)

	resultResp, stop := Handler4(req, resp)
	assert.False(t, stop) // Should not stop processing
	assert.NotNil(t, resultResp)

	// Create a DHCPv4 request with a non-whitelisted MAC
	mac2, err := net.ParseMAC("00:00:00:00:00:00")
	assert.NoError(t, err)
	req2, err := dhcpv4.NewDiscovery(mac2)
	assert.NoError(t, err)

	// Process the request - should be blocked
	resp2, err := dhcpv4.NewReplyFromRequest(req2)
	assert.NoError(t, err)

	resultResp2, stop2 := Handler4(req2, resp2)
	assert.True(t, stop2) // Should stop processing (drop request)
	assert.Nil(t, resultResp2)

	// Test with empty whitelist (should allow all)
	whitelistedMACs = make(map[string]bool) // Clear whitelist
	resp3, err := dhcpv4.NewReplyFromRequest(req)
	assert.NoError(t, err)

	resultResp3, stop3 := Handler4(req, resp3)
	assert.False(t, stop3) // Should not stop processing
	assert.NotNil(t, resultResp3)
}

func TestWhitelistHandler6(t *testing.T) {
	// Setup whitelist with specific MAC addresses
	_, err := setup6("00:11:22:33:44:55", "aa:bb:cc:dd:ee:ff")
	assert.NoError(t, err)

	// Note: Testing DHCPv6 handler is more complex due to the nature of DHCPv6 packets
	// and MAC extraction. In a real scenario, we would need to mock a complete DHCPv6
	// exchange to properly test this.
}
