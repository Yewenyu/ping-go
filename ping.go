package goPing

import (
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-ping/ping"
)

type PingResults struct {
	Host string  `json:"host"`
	Port int     `json:"port"`
	Time int     `json:"time"`
	Loss float64 `json:"loss"`
}

func StartPings(hosts []string, pingCount int) map[string]PingResults {
	var wg sync.WaitGroup
	count := len(hosts)
	wg.Add(count)

	var lock sync.Mutex
	var resultm = map[string]PingResults{}

	for _, host := range hosts {
		go func(host string) {
			state := StartPing(host, pingCount)
			if state != nil {
				loss := state.PacketLoss
				timev := state.AvgRtt
				result := PingResults{
					Host: host,
					Time: int(timev.Milliseconds()),
					Loss: loss,
				}
				lock.Lock()
				resultm[host] = result
				lock.Unlock()
			}
			wg.Done()
			debug.FreeOSMemory()
		}(host)
	}

	wg.Wait()
	return resultm
}

func StartPing(host string, pingCount int) *ping.Statistics {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return nil
	}
	pinger.Count = pingCount
	pinger.Timeout = 2 * time.Second
	pinger.OnRecv = func(pkt *ping.Packet) {
		// fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
		// 	pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
	}
	pinger.OnFinish = func(stats *ping.Statistics) {
		// fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
		// fmt.Printf("%d packets transmitted, %d packets received, %v%% packet loss\n",
		// 	stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		// fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
		// 	stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
	}
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		return nil
	}
	stats := pinger.Statistics()

	return stats
}
