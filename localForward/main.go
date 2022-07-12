package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
)

func tcpListen(local,target int){
    listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d",local))
    if err != nil {
        fmt.Println("listen failed, err:", err)
        return
    }
 
    fmt.Println("listen Start...:")
 
    for {
        //2.接收客户端的链接
        client, err := listen.Accept()
        if err != nil {
            fmt.Printf("accept failed, err:%v\n", err)
            continue
        }
        go func(){
            server, err := net.Dial("tcp", fmt.Sprintf("0.0.0.0:%d",target))
            defer func(){
                client.Close()
                if server != nil{
                    server.Close()
                }
            }()
            if err != nil {
                fmt.Printf("dial failed, err:%v\n", err)
                return
            }
            go io.Copy(client,server)
            io.Copy(server,client)
        }()
    }
}
func udpListen(local,target int){
    listen, err := net.ListenUDP("udp", &net.UDPAddr{
        IP:   net.IPv4(0, 0, 0, 0),
        Port: local,
    })
    if err != nil {
        fmt.Printf("listen failed, err:%v\n", err)
        return
    }
    socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
        IP:   net.IPv4(127, 0, 0, 1),
        Port: target,
    })
    if err != nil {
        fmt.Println("连接失败!", err)
        return
    }
    go func(){
        islisten := false
        for{
            var data [1024]byte
            //读取UDP数据
            count, addr, err := listen.ReadFromUDP(data[:])
            if err != nil {
                fmt.Printf("read udp failed, err:%v\n", err)
                continue
            }
            _,_ = socket.Write(data[:count])
            if !islisten{
                islisten = true
                go func(){
                    for{
                        var data [1024]byte
            //读取UDP数据
                        n,err := socket.Read(data[:])
                        if err != nil{
                            continue
                        }
                        listen.WriteToUDP(data[:n],addr)
                    }
                }()
            }
        }
    }()
    
}
func main(){

    localPort := 6666
    targetPort := 7778

	go tcpListen(localPort,targetPort)
    // go tcpListen(6667,7779)

    go udpListen(localPort,targetPort)
    // go udpListen(6667,1234)

	http.ListenAndServe("0.0.0.0:6060", nil)
}