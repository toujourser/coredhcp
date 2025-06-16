package next_server

import (
	"fmt"
	"net"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/toujourser/coredhcp/handler"
	"github.com/toujourser/coredhcp/logger"
	"github.com/toujourser/coredhcp/plugins"
)

// PluginName is the name of the plugin
const PluginName = "next_server"

// Logger for the next-server plugin
var log = logger.GetLogger("plugins/next_server")

// Plugin is the next-server plugin registration
var Plugin = plugins.Plugin{
	Name:   PluginName,
	Setup4: setup4,
	Setup6: nil, // No DHCPv6 support for next-server
}

// NextServerPlugin holds the plugin configuration
type NextServerPlugin struct {
	NextServer net.IP // IP address of the next server (TFTP server)
}

// setup4 initializes the next-server plugin for DHCPv4
func setup4(args ...string) (handler.Handler4, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("next_server plugin requires an IP address argument")
	}
	ip := net.ParseIP(args[0])
	if ip == nil || ip.To4() == nil {
		return nil, fmt.Errorf("invalid IPv4 address for next_server: %s", args[0])
	}
	log.Infof("next_server plugin loaded with IP: %s", ip.String())
	return NextServerPlugin{NextServer: ip}.handle4, nil
}

// handle4 processes DHCPv4 packets and sets the next-server (siaddr) field
func (p NextServerPlugin) handle4(req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
	if resp == nil {
		return nil, true // Drop if no response
	}
	// Set the next-server (siaddr) field in the DHCPv4 response
	resp.ServerIPAddr = p.NextServer.To4()
	log.Debugf("Set next-server %s in DHCP response for client %s", p.NextServer.String(), req.ClientHWAddr.String())
	return resp, false // Continue to next plugin
}
