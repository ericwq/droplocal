package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

func getIPv4from(hostname string) (string, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return "", err
	}

	for _, ip := range ips {
		//this is the first way to distinguish ipv4 and ipv6 address
		if ip.To4() != nil {
			//the first ipv4 address is enough for us. ignore the others
			return ip.String(), nil
		}
	}
	return "", errors.New("can't find ipv4 address")
}

func instanceNameFactory() string {
	//return servNamePrefix

	//	hostname, err := os.Hostname()
	//		if err != nil {
	//			log.Fatalf("get hostname error: %s", err)
	//		}
	// this one doesn't work
	//	hostname := "qiwang.local"
	//var id unit
	rand.Seed(time.Now().UnixNano())

	//TODO the random id still has the potential to conflict in crowded network.
	id := rand.Uint32()
	return fmt.Sprintf("%s %d", servNamePrefix, id)

}

func filterInterface() []net.Interface {

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("net.Interface() err:%s", err)
	}

	var nif []net.Interface
	for _, val := range interfaces {

		switch {
		case val.Flags&net.FlagLoopback != 0:
			//it's loop back
			continue
		case val.Flags&net.FlagPointToPoint != 0:
			//it's p2p
			continue
		case val.Flags&net.FlagUp == 0:
			//it's NOT up
			continue
		default:
			nif = append(nif, val)
		}
	}

	return nif
}
