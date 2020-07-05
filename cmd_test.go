package main

import (
	"fmt"
	"net"
	"testing"
)

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
