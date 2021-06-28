package main

/*
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	_ "net/http/pprof"

	goPing "github.com/Yewenyu/ping-go"
)

//export startPing
func startPing(hosts *C.char, pingCount int) *C.char {
	ss := C.GoString(hosts)
	var h = strings.Split(ss, ",")
	statam := goPing.StartPings(h, pingCount)
	if len(statam) == 0 {
		return nil
	}

	mapv := make(map[string]interface{})
	for key, v := range statam {
		jsonStr, err := json.Marshal(v)
		if err == nil {
			mapv[key] = string(jsonStr)
		}

	}
	str, e := json.Marshal(mapv)
	if e != nil {
		return nil
	}
	var s = string(str)
	debug.FreeOSMemory()
	return C.CString(s)
}

//export startTcpPing
func startTcpPing(addrs *C.char, pingCount int, timeout int, interval int) *C.char {
	ss := C.GoString(addrs)
	var h = strings.Split(ss, ",")
	result := goPing.StartTcpPings(h, pingCount, timeout, interval)
	mapv := make(map[string]interface{})
	for _, r := range result {
		jsonStr, err := json.Marshal(r)
		if err == nil {
			mapv[r.Host] = string(jsonStr)
		}

	}
	str, e := json.Marshal(mapv)
	if e != nil {
		return nil
	}
	var s = string(str)
	debug.FreeOSMemory()
	return C.CString(s)
}
func main() {

	// ss := goPing.StartPings([]string{"www.baidu.com", "www.qq.com"}, 3)
	ss := goPing.StartTcpPings([]string{"www.baidu.com:443", "www.qq.com:443"}, 3, 2, 1)
	fmt.Printf("%s\n", ss)

	go func() {
		for {

			time.Sleep(1 * time.Second)
		}
	}()
	http.ListenAndServe("0.0.0.0:6060", nil)
}
