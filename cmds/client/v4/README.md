# DHCPv4 debug client

This is a simple dhcpv4 client for use as a debugging tool with coredhcp

***This is not a general-purpose DHCP client. This is only a testing/debugging tool for developing CoreDHCP***

## Usage

```bash
# Basic usage with broadcast (default)
./dhcpv4-client [MAC-ADDRESS]

# Specify server and local addresses
./dhcpv4-client -server <SERVER_IP> -local <LOCAL_IP> [MAC-ADDRESS]

# Examples:
# Use broadcast to find DHCP server on local network
./dhcpv4-client -server 255.255.255.255 -local 0.0.0.0 fa:16:3e:ac:e6:e6

# Connect to specific DHCP server
./dhcpv4-client -server 192.168.1.1 -local 192.168.1.100 fa:16:3e:ac:e6:e6

# Loopback testing (server must listen on 127.0.0.1)
./dhcpv4-client -server 127.0.0.1 -local 127.0.0.1 fa:16:3e:ac:e6:e6
```

## Flags

- `-server`: DHCP server IP address (default: `255.255.255.255` for broadcast)
- `-local`: Local IP address to bind to (default: `0.0.0.0` to listen on all interfaces)
- First argument: MAC address (default: `fa:16:3e:ac:e6:e5`)
