# MAC Address Whitelist Plugin

## Overview

The whitelist plugin allows you to restrict DHCP IP assignments to only specific MAC addresses. Only clients with MAC addresses listed in the whitelist configuration will be able to obtain IP addresses from the DHCP server.

## Features

- Supports both DHCPv4 and DHCPv6
- Configurable through the main configuration file
- If no whitelist is configured, there are no restrictions (allows all MAC addresses)
- Case-insensitive MAC address matching
- Detailed logging for monitoring and debugging

## Configuration

To use the whitelist plugin, add it to your CoreDHCP configuration file:

```yaml
server6:
    listen: '[::]:547'
    plugins:
        - whitelist:
            - "00:11:22:33:44:55"
            - "aa:bb:cc:dd:ee:ff"
        - server_id: LL 00:de:ad:be:ef:00
        - file: "leases.txt"

server4:
    listen: '0.0.0.0:67'
    plugins:
        - whitelist:
            - "00:11:22:33:44:55"
            - "aa:bb:cc:dd:ee:ff"
        - server_id: 10.0.0.1
        - range: "leases.txt" "10.0.0.100" "10.0.0.200" "24h"
```

## Behavior

1. **With whitelist configured**: Only clients with MAC addresses in the whitelist can obtain IP assignments. Other requests are dropped.
2. **Without whitelist configured**: All clients can obtain IP assignments (no restrictions).
3. **Empty whitelist**: Same as no whitelist configured - all clients can obtain IP assignments.

## Logging

The plugin logs the following events:

- When a MAC address is found in the whitelist (allowed)
- When a MAC address is not found in the whitelist (dropped)
- When MAC address extraction fails

Logs are written using the standard CoreDHCP logger with the prefix "plugins/whitelist".

## Testing

Run the plugin tests with:

```bash
go test ./plugins/whitelist/...
```