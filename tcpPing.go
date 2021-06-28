package goPing

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	ping "github.com/cloverstd/tcping/ping"
)

func StartTcpPings(addrs []string, pingCount int, timeout int, interval int) []PingResults {
	var wg sync.WaitGroup
	count := len(addrs)
	wg.Add(count)

	var lock sync.Mutex
	var resultm = make([]PingResults, count)

	for i, addr := range addrs {
		arr := strings.Split(addr, ":")
		host := arr[0]
		port, err := strconv.Atoi(arr[1])
		if err != nil {
			continue
		}
		go func(i int) {
			r := StartTcpPing(host, port, pingCount, timeout, interval)
			lock.Lock()
			resultm[i] = r
			lock.Unlock()
			wg.Done()
			debug.FreeOSMemory()
		}(i)

	}
	wg.Wait()
	return resultm
}
func StartTcpPing(host string, port int, pingCount int, timeout int, interval int) PingResults {

	target := ping.Target{
		Timeout:  time.Duration(timeout) * time.Second,
		Interval: time.Duration(interval) * time.Second,
		Host:     host,
		Port:     port,
		Counter:  pingCount,
		Protocol: ping.TCP,
	}
	pinger := ping.NewTCPing()
	pinger.SetTarget(&target)
	pingerDone := pinger.Start()
	select {
	case <-pingerDone:
		break
	}
	result := pinger.Result()

	fmt.Println(result)
	r := PingResults{
		Host: host,
		Port: port,
		Time: int(result.Avg().Milliseconds()),
		Loss: (1 - float64(result.SuccessCounter)/float64(result.Counter)) * 100,
	}
	return r
}
