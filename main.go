package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

var (
	debug      bool = false
	r               = &net.Resolver{}
	nat64      string
	nameserver string
	/*
		Regexs for checking input
	*/
	ipv6Regex = `^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`
	ipv4Regex = `^(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4})`
)

func isMayOnlyIPv6(host string) bool {
	match, err := regexp.MatchString(ipv6Regex, host)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return match
}

func isMayOnlyIPv4(host string) bool {
	match, err := regexp.MatchString(ipv4Regex, host)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return match
}

func parseQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeAAAA:
			ips, err := r.LookupIP(context.Background(), "ip4", q.Name)
			if debug {
				log.Printf("AAAA Query for %s\n", q.Name)
			}

			if err == nil {
				for _, ip := range ips {
					ip_respond := ip.String()
					if strings.Contains(ip_respond, ".") {
						if debug {
							log.Printf("A response from %s for %s %s\n", "1.1.1.1", q.Name, ip_respond)
						}
						rr, err := dns.NewRR(fmt.Sprintf("%s AAAA %s", q.Name, nat64+ip_respond))
						if err == nil {
							m.Answer = append(m.Answer, rr)
						}
					}

				}
			} else {
				log.Printf("Resolve err for \"%s\" : %s", q.Name, err)
			}

		}
	}
}

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		parseQuery(m)
	}

	w.WriteMsg(m)
}

func init() {
	flag.StringVar(&nameserver, "nameserver", "1.1.1.1", "Define your nameserver")
	flag.StringVar(&nat64, "nat64", "64:ff9b::", "NAT64 IPv6 prefix")
	flag.BoolVar(&debug, "D", false, "Debug mode")
	flag.Parse()
}

func main() {
	if isMayOnlyIPv4(nameserver) {
		nameserver = nameserver + ":53"
	} else if isMayOnlyIPv6(nameserver) {
		nameserver = "[" + nameserver + "]:53"
	} else {
		log.Fatalf("Nameserver parse error %s\n", nameserver)
	}
	r = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, nameserver)
		},
	}
	log.Printf("Nameserver is %s\n", nameserver)
	log.Printf("Nat64 Prefix is %s\n", nat64)
	// attach request handler func
	dns.HandleFunc(".", handleDnsRequest)
	// start server
	port := 53
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	log.Printf("NAT64 only DNS server starting at %d\n", port)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start DNS server: %s\n ", err.Error())
	}
}
