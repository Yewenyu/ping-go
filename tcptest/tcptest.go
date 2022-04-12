package tcptest

import (
	"context"
	"io"
	"net"
	"sync"
	"time"
)

type Result struct {
	Target   string
	Duration time.Duration
}


func CheckTCPAlive(ctx context.Context, addrs []string) (alive []Result, dead []Result, timeout []Result) {
	alive = make([]Result, 0, len(addrs))
	var aliveLock sync.Mutex
	dead = make([]Result, 0, len(addrs))
	var deadLock sync.Mutex
	timeout = make([]Result, 0, len(addrs))
	var timeoutLock sync.Mutex
	var wg sync.WaitGroup
	for _, addr := range addrs {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			success, dura := CheckTCP64(ctx, addr)
			if success {
				aliveLock.Lock()
				alive = append(alive, Result{addr, dura})
				aliveLock.Unlock()
			} else {
				if ctx.Err() != nil {
					timeoutLock.Lock()
					timeout = append(timeout, Result{addr, dura})
					timeoutLock.Unlock()
				} else {
					deadLock.Lock()
					dead = append(dead, Result{addr, dura})
					deadLock.Unlock()
				}
			}
		}(addr)
	}
	wg.Wait()
	return
}

var Send64Bytes = true
func CheckTCP64(ctx context.Context, addr string) (bool, time.Duration) {
	start := time.Now()
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		dura := time.Since(start)
		return false, dura
	}
	defer conn.Close()
	if !Send64Bytes{
		return true,time.Since(start)
	}
	newCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func(ctx context.Context) {
		<-ctx.Done()
		conn.SetDeadline(time.Now())
	}(newCtx)
	buf := make([]byte, 64)
	n, err := conn.Write(buf)
	if err != nil {
		dura := time.Since(start)
		return false, dura
	}
	if n != 64 {
		dura := time.Since(start)
		return false, dura
	}
	for recvLen := 0; recvLen < 64; {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			dura := time.Since(start)
			return false, dura
		}
		recvLen += n
	}
	dura := time.Since(start)
	return true, dura
}