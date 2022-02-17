package ping

import (
	"errors"
	"fmt"
	"net"
	"time"
)

// TCPing ...
type TCPing struct {
	target     *Target
	done       chan struct{}
	result     *Result
	CheckAlive bool
	ErrType int
}

var _ Pinger = (*TCPing)(nil)

// NewTCPing return a new TCPing
func NewTCPing() *TCPing {
	tcping := TCPing{
		done: make(chan struct{}),
	}
	return &tcping
}

// SetTarget set target for TCPing
func (tcping *TCPing) SetTarget(target *Target) {
	tcping.target = target
	if tcping.result == nil {
		tcping.result = &Result{Target: target}
	}
}

// Result return the result
func (tcping TCPing) Result() *Result {
	return tcping.result
}

// Start a tcping
func (tcping *TCPing) Start() <-chan struct{} {
	go func() {
		t := time.NewTicker(tcping.target.Interval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				if tcping.result.Counter >= tcping.target.Counter && tcping.target.Counter != 0 {
					tcping.Stop()
					return
				}
				duration, _, err := tcping.ping()
				tcping.result.Counter++

				if err != nil {
					// fmt.Printf("Ping %s - failed: %s\n", tcping.target, err)
				} else {
					// fmt.Printf("Ping %s(%s) - Connected - time=%s\n", tcping.target, remoteAddr, duration)

					if tcping.result.MinDuration == 0 {
						tcping.result.MinDuration = duration
					}
					if tcping.result.MaxDuration == 0 {
						tcping.result.MaxDuration = duration
					}
					tcping.result.SuccessCounter++
					if duration > tcping.result.MaxDuration {
						tcping.result.MaxDuration = duration
					} else if duration < tcping.result.MinDuration {
						tcping.result.MinDuration = duration
					}
					tcping.result.TotalDuration += duration
				}
			case <-tcping.done:
				return
			}
		}
	}()
	return tcping.done
}

// Stop the tcping
func (tcping *TCPing) Stop() {
	tcping.done <- struct{}{}
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(tcping.done)
	}()
}

func (tcping *TCPing) ping() (time.Duration, net.Addr, error) {
	var remoteAddr net.Addr
	duration, errIfce := timeIt(func() interface{} {
		conn, err := net.DialTimeout("tcp",fmt.Sprintf("%s:%d", tcping.target.Host, tcping.target.Port),tcping.target.Timeout)
		if err != nil {
			tcping.ErrType = 1
			return err
		}
		if tcping.CheckAlive {
			conn.SetReadDeadline(time.Now().Add(tcping.target.Timeout))
			
			buf := make([]byte, 100)
			buf1 := make([]byte,64)
			go func(){
				for{
					_,err := conn.Write(buf1)
					if err != nil{
						break
					}
					time.Sleep(1 * time.Second)
				}
			}()
			
			// c := conn.(*net.TCPConn)
			n, err := conn.Read(buf)
			if err != nil {
				tcping.ErrType = 2
				return err
			}
			if n == 5{
				fmt.Print(tcping.target.Host + "\n")
			}
			
			if n == 64 || n == 5{
				
			}else{
				tcping.ErrType = 3
				s := string(buf[0:n])
				_ = s
				return errors.New("not alive")
			}
		}

		remoteAddr = conn.RemoteAddr()
		conn.Close()
		return nil
	})
	if errIfce != nil {
		err := errIfce.(error)
		return 0, remoteAddr, err
	}
	return time.Duration(duration), remoteAddr, nil
}
