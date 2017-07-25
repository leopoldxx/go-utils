package net

import (
	"context"
	"encoding/binary"
	"net"
	"sync"
	"syscall"
	"time"

	"gopkg.in/vmihailenco/msgpack.v2"

	"github.com/leopoldxx/go-utils/concurrency"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	protocolICMP  = 1
	concurrentNum = 100
	maxRetryTimes = 3
	retryQueueLen = 1000
)

type payload struct {
	Start int64
}

// Status of the target host
type Status struct {
	OK  bool
	RTT time.Duration
}

type pinger struct {
	ttl time.Duration
}

func sendEchoMessage(conn *icmp.PacketConn, wg *sync.WaitGroup, cb *concurrency.Barrier, ip net.IP) {
	if conn == nil || ip == nil {
		return
	}
	if wg != nil {
		defer wg.Done()
	}
	if cb != nil {
		defer cb.Done()
	}

	pl := payload{
		Start: time.Now().UnixNano(),
	}

	b, _ := msgpack.Marshal(&pl)

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   int(binary.BigEndian.Uint16(ip[2:])),
			Seq:  int(binary.BigEndian.Uint16(ip[0:2])),
			Data: b,
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		//TODO: log
		return
	}

	addr := &net.IPAddr{IP: ip}
	for {
		if _, err := conn.WriteTo(wb, addr); err != nil {
			if neterr, ok := err.(*net.OpError); ok {
				if neterr.Err == syscall.ENOBUFS {
					continue
				}
			}
		}
		break
	}
}

func (p pinger) sendEchoMessages(ctx context.Context, wg *sync.WaitGroup, cb *concurrency.Barrier, conn *icmp.PacketConn, ips []string) {
	for i := range ips {
		select {
		case <-ctx.Done():
			return
		default:
		}

		cb.Advance()
		wg.Add(1)
		go sendEchoMessage(conn, wg, cb, net.ParseIP(ips[i]).To4())
	}
	wg.Wait()
}
func (p pinger) recvEchoReplyMessage(ctx context.Context, cb *concurrency.Barrier, conn *icmp.PacketConn, ips []string) (map[string]Status, error) {
	result := map[string]Status{}
	for i := range ips {
		result[ips[i]] = Status{OK: false}
	}
	retryIPs := map[string]int{}
	ipNum := len(ips)
	pingNum := 0

	newCtx, cancel := context.WithTimeout(ctx, p.ttl)
	conn.SetReadDeadline(time.Now().Add(p.ttl))
	defer cancel()

	retryCh := make(chan string, retryQueueLen)
	defer close(retryCh)

	go func(retry <-chan string) {
		for ip := range retry {
			cb.Advance()
			go sendEchoMessage(conn, nil, cb, net.ParseIP(ip).To4())
		}
	}(retryCh)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-newCtx.Done():
			return result, nil
		default:
		}

		rb := make([]byte, 256)
		n, peer, err := conn.ReadFrom(rb)
		if err != nil {
			//conn.SetReadDeadline(time.Now().Add(timeout))
			continue
		}
		rm, err := icmp.ParseMessage(protocolICMP, rb[:n])
		if neterr, ok := err.(*net.OpError); ok {
			if neterr.Timeout() {
				// TODO: Read Timeout
			}
			continue
		}
		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			if status, exists := result[peer.String()]; exists {
				if !status.OK {
					var startTime *int64
					if pkt, ok := rm.Body.(*icmp.Echo); ok {
						//if pkt.ID == p.id && pkt.Seq == p.seq {
						//	            rtt = time.Since(bytesToTime(pkt.Data[:TimeSliceLength]))
						//				        }

						pl := payload{}
						err = msgpack.Unmarshal(pkt.Data, &pl)
						if err == nil {
							startTime = &pl.Start
						}
					}
					now := time.Now().UnixNano()
					var diff int64 = -1
					if startTime != nil && now > *startTime {
						diff = now - *startTime
					}

					result[peer.String()] = Status{OK: true, RTT: time.Duration(diff)}
					pingNum++
				}
			}
			if pingNum == ipNum {
				return result, nil
			}
		default:
			num := retryIPs[peer.String()]
			if num < maxRetryTimes {
				retryIPs[peer.String()] = num + 1
				retryCh <- peer.String()
			}
		}
	}
}

func (p pinger) Ping(ctx context.Context, ips []string) (map[string]Status, error) {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}
	defer c.Close()

	wg := &sync.WaitGroup{}
	cb := concurrency.NewBarrier(concurrentNum)

	go p.sendEchoMessages(ctx, wg, cb, c, ips)
	return p.recvEchoReplyMessage(ctx, cb, c, ips)
}

type option struct {
	ttl time.Duration
}

type Option func(opt *option)

func WithTTL(ttl time.Duration) Option {
	return func(opt *option) {
		opt.ttl = ttl
	}
}

// Ping do ping
func Ping(ctx context.Context, ips []string, opts ...Option) (map[string]Status, error) {
	opt := &option{
		ttl: time.Second * 2,
	}
	for _, ops := range opts {
		ops(opt)
	}
	return pinger{
		ttl: opt.ttl,
	}.Ping(ctx, ips)
}
