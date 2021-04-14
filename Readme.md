# DNS64-Only

Test your NAT64 network with DNS64-Only

## Avaible Options

- **nameserver**  
Set upstream dns server.  
Default: 1.1.1.1

- **D**  
Debug mode.

- **nat64**  
IPv6 nat64 prefix.  
Default: 64:ff9b::

## Example Run

You can start dns64-only with docker or if you prefer you can build from source.

```bash
docker run -it --rm ghcr.io/ahmetozer/dns64-only

# Custom Nameserver
docker run -it --rm ghcr.io/ahmetozer/dns64-only -nameserver 8.8.8.8
```

## Work Logic

The server reads DNS queries and asks upstream DNS server. After getting the response, the system translates IPv4 addresses with your NAT64 IPv6 prefix and only reply the client request with the translated address.

![dns64-only-work-logic](dns64-only.png)
