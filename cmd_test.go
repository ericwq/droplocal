package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
)

func TestGetIPv4From(t *testing.T) {
	//hostname, _ := os.Hostname()
	//getIPv4from()
	hostname, _ := os.Hostname()
	addr, _ := net.LookupIP(hostname) //"Oppenheimer.local")
	for _, a := range addr {
		//this is the first way to distinguish ipv4 and ipv6 address
		if a.To4() != nil {
			fmt.Println("it's ipv4", a)
		} else {
			fmt.Println("it's ipv6", a)
		}
		//this is the second way to distinguish ipv4 and ipv6 address
		if strings.Contains(a.String(), ":") {
			fmt.Println("IPv6 address: ", a)
		} else {

			fmt.Println("IPv4 address: ", a)
		}

	}
}

func TestIF(t *testing.T) {
	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		t.Fatalf("net.Interface() err:%s", err)
	}

	var nif []net.Interface
	for _, val := range interfaces {

		//fmt.Printf("%d : %v\n", k, i)
		switch {
		case val.Flags&net.FlagLoopback != 0:
			//it's loop back
			continue
		case val.Flags&net.FlagPointToPoint != 0:
			//it's p2p
			continue
		case val.Flags&net.FlagUp == 0:
			//it's not up
			continue
		default:
			nif = append(nif, val)
			fmt.Printf("add: %v\n", val)
		}
	}

	for i, v := range nif {
		fmt.Printf("%d: %v\n", i, v)
	}
}
