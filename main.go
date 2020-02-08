package main

import (
        "fmt"
        "net"
        "os"
        "strings"
        "github.com/nadoo/conflag"
        "bufio"
        "strconv"
        "time"
)

// conflag
var conf struct {
    isServer bool
    isClient bool
    LocalAddress string
    RemoteAddress string
}

func main() {
        flag := conflag.New()//生成conflag实例

        flag.BoolVar(&conf.isServer,"s",false,"Whether it is a server")
        flag.BoolVar(&conf.isClient,"c",false,"Whether it is a client")
        flag.StringVar(&conf.LocalAddress,"l","0.0.0.0:9000","Local Listen Address")
        flag.StringVar(&conf.RemoteAddress,"r","127.0.0.1:9000","Remote Server Address")

        flag.Parse()//解析命令行参数

        fmt.Printf("%+v\n",conf)
        if conf.isServer && conf.isClient {
                fmt.Println("It is not allowed to run server and client at the same time!")
        }else if conf.isServer && !conf.isClient {
                runAsServer(conf.LocalAddress)
        }else if !conf.isServer && conf.isClient {
                runAsClient(conf.RemoteAddress)
        }else{
                fmt.Println("No operating mode specified!")
        }
}
// runAsClient
// 作为客户端运行
func runAsClient(CONNECT string) {
        srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 1999} // 注意端口必须固定
        s, err := net.ResolveUDPAddr("udp4", CONNECT)
        c, err := net.DialUDP("udp4", srcAddr, s)
        if err != nil {
                fmt.Println(err)
                return
        }

        fmt.Printf(timeDate() + "The UDP server is %s\n", c.RemoteAddr().String())

        paired,connected := false, false
        var peer net.UDPAddr

        for {
                reader := bufio.NewReader(os.Stdin)
                fmt.Print(">> ")
                text, _ := reader.ReadString('\n')
                data := []byte(text + "\n")
                _, err = c.Write(data)
                if strings.TrimSpace(string(data)) == "STOP" {
                        fmt.Println("Exiting UDP client!")
                        return
                }

                if err != nil {
                        fmt.Println(err)
                        return
                }

                buffer := make([]byte, 1024)
                n, _, err := c.ReadFromUDP(buffer)
                if err != nil {
                        fmt.Println(err)
                        return
                }

                if(!paired){
                        paired = true
                        peer = parseAddr(string(buffer[0:n]))
                        fmt.Printf(timeDate() + "已获取对端地址：%s\n",string(buffer[0:n]))
                        c.Close()
                }
                
                break
        }

        fmt.Println(timeDate() + "打洞前等待")
        if(paired && !connected){
                //进行打洞
                fmt.Println(timeDate() + "打洞开始")
                conn, err := net.DialUDP("udp", srcAddr, &peer)
                if err != nil {
                        fmt.Println(err)
                }
                connected = true

                defer conn.Close()
                // 向另一个peer发送一条udp消息(对方peer的nat设备会丢弃该消息,非法来源),用意是在自身的nat设备打开一条可进入的通道,这样对方peer就可以发过来udp消息
                if _, err = conn.Write([]byte("peer")); err != nil {
                        fmt.Println(timeDate() + "send handshake:", err)
                }
                // 心跳包
                go func() {
                        for {
                                time.Sleep(3 * time.Second)
                                data := []byte("ping")
                                if _, err = conn.Write(data); err != nil {
                                        fmt.Println(timeDate() + "ping fail", err)
                                }
                        }
                }()
                go func() {
                        for {
                                //time.Sleep(10 * time.Second)
                                reader := bufio.NewReader(os.Stdin)
                                text, _ := reader.ReadString('\n')
                                data := []byte(strings.Replace(text, "\n", "", -1))//去除回车符
                                if _, err = conn.Write(data); err != nil {
                                        fmt.Println(timeDate() + "send msg fail", err)
                                }
                        }
                }()
                for {
                        data := make([]byte, 1024)
                        n, _, err := conn.ReadFromUDP(data)
                        if err != nil {
                                fmt.Printf(timeDate() + "error during read: %s\n", err)
                        } else {
                                if stringCompare(data[:n], []byte("ping")) != 0 {
                                        fmt.Printf("%s\n", data[:n])
                                }
                        }
                }
        }

}
// runAsServer
// 作为服务端运行
func runAsServer(PORT string) {
        s, err := net.ResolveUDPAddr("udp4", PORT)
        if err != nil {
                fmt.Println(err)
                return
        }

        connection, err := net.ListenUDP("udp4", s)
        if err != nil {
                fmt.Println(err)
                return
        }

        defer connection.Close()
        buffer := make([]byte, 1024)

        clients := make([]net.UDPAddr, 0, 2)
        clients_string := make([]string, 0, 2)
        for {
                n, addr, err := connection.ReadFromUDP(buffer)
                if err != nil {
                        fmt.Println(err)
                        return
                }
                
                uid := -1

                if uid == -1 {
                        clients = append(clients, *addr)
                        //fmt.Printf("%+v\n", clients)
                        clients_string = append(clients_string, addr.String())
                        //fmt.Printf("%+v\n", clients_string)
                }

                for i,v := range clients_string {
                        if addr.String() == v {
                                uid = i
                        }
                }

                fmt.Print(timeDate() + "-> ", string(buffer[0:n-1]))

                if strings.TrimSpace(string(buffer[0:n])) == "STOP" {
                        fmt.Println(timeDate() + "Exiting UDP server!")
                        return
                }

                if len(clients) == 2 {
                        connection.WriteToUDP([]byte(clients[0].String()), &clients[1])
                        connection.WriteToUDP([]byte(clients[1].String()), &clients[0])
                        clients_string = clients_string[:0]//清空
                        clients = clients[:0]//清空
                }

        }
}
func stringCompare(a, b[]byte) int {
    for i:=0; i < len(a) && i < len(b); i++ {
        switch {
        case a[i] > b[i]:
            return 1
        case a[i] < b[i]:
            return -1
        }
    }
    // 数组的长度可能不同
    switch {
    case len(a) < len(b):
        return -1
    case len(a) > len(b):
        return 1
    }
    return 0 // 数组相等
}
func parseAddr(addr string) net.UDPAddr {
        t := strings.Split(addr, ":")
        port, _ := strconv.Atoi(t[1])
        return net.UDPAddr{
                IP:   net.ParseIP(t[0]),
                Port: port,
        }
}
// timeDate
// 返回时间
func timeDate() string {
    return time.Now().Format("2006-01-02 15:04:05")
}
