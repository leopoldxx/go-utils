package net_test

import (
	"context"
	"log"
	"time"

	"github.com/leopoldxx/go-utils/net"
)

func ExamplePing() {
	targetIP := []string{
		"8.8.8.8",
		"1.1.1.1",
		"2.2.2.2",
		"61.135.169.121",
		"61.135.157.156",
		"123.126.104.68",
		"211.144.7.85",
	}

	log.SetFlags(log.Lmicroseconds | log.LstdFlags)
	log.Printf("ping")
	res, err := net.Ping(context.TODO(), targetIP, net.WithTTL(time.Millisecond*800))
	if err != nil {
		log.Printf("ping failed: %s", err)
		return
	}
	for ip, status := range res {
		log.Printf("%s is %s, %v", ip, func(status net.Status) string {
			if status.OK == true {
				return "ok"
			}
			return "unreachable"
		}(status), status.RTT)
	}
	// Output:
}
