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
        "unicode/utf8"
)

// conflag
var conf struct {
    isServer bool
    isClient bool
    LocalAddress string
    LocalKey string
    PairedKey string
    RemoteAddress string
}

func main() {
        flag := conflag.New()//生成conflag实例

        flag.BoolVar(&conf.isServer,"s",false,"Whether it is a server")
        flag.BoolVar(&conf.isClient,"c",false,"Whether it is a client")
        flag.StringVar(&conf.LocalAddress,"l","0.0.0.0:9000","Local Listen Address")
        flag.StringVar(&conf.LocalKey,"lk","99998","Local Key")
        flag.StringVar(&conf.PairedKey,"pk","99999","Paired Key")
        flag.StringVar(&conf.RemoteAddress,"r","127.0.0.1:9000","Remote Server Address")

        flag.Parse()//解析命令行参数

        fmt.Printf("%+v\n",conf)
        if conf.isServer && conf.isClient {
                fmt.Println("It is not allowed to run server and client at the same time!")
        }else if conf.isServer && !conf.isClient {
                runAsServer(conf.LocalAddress)
        }else if !conf.isServer && conf.isClient {
                runAsClient(conf.RemoteAddress, conf.LocalKey, conf.PairedKey)
        }else{
                fmt.Println("No operating mode specified!")
        }
}
// runAsClient
// 作为客户端运行
func runAsClient(CONNECT, LOCALKEY, PAIREDKEY string) {
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
                data := LOCALKEY + "|" + PAIREDKEY
                pairedKeyLength := utf8.RuneCountInString(data) - 1
                if pairedKeyLength > 32 && pairedKeyLength < 16 {
                        fmt.Println("PAIREDKEY not allowed to be too long")
                        return
                }
                _, err = c.Write([]byte(data))

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

        clients := make([]net.UDPAddr, 0, 20)//最多同时20个客户端
        clients_obj := make([]map[string]string, 0, 20)//最多同时20个客户端
        for {
                n, addr, err := connection.ReadFromUDP(buffer)
                if err != nil {
                        fmt.Println(err)
                        return
                }

                data := strings.Split(string(buffer[0:n]), "|")

                ulid,upid := -1, -1
                for i,v := range clients_obj {
                        if data[1] == v["localKey"] {
                                upid = i
                        }
                        if addr.String() == v["clientAddr"] {
                                ulid = i
                        }
                }

                if ulid == -1 {
                        clients = append(clients, *addr)
                        //fmt.Printf("%+v\n", clients)
                        client := make(map[string]string)
                        client["localKey"] = data[0]
                        client["localId"] = fmt.Sprintf("%d", ulid)
                        client["pairedKey"] = data[1]
                        client["pairedId"] = fmt.Sprintf("%d", upid)
                        client["clientAddr"] = addr.String()
                        clients_obj = append(clients_obj, client)
                        client["localId"] = fmt.Sprintf("%d", len(clients_obj)-1)
                        ulid = len(clients_obj)-1
                        //fmt.Printf("%+v\n", clients_obj)
                }

                fmt.Printf(timeDate() + "-> (%d)LocalKey: %s | (%d)pairedKey: %s\n", ulid, data[0], upid, data[1])

                if ulid != -1 && upid != -1 {
                        connection.WriteToUDP([]byte(clients[ulid].String()), &clients[upid])
                        connection.WriteToUDP([]byte(clients[upid].String()), &clients[ulid])
                        clients_obj = append(clients_obj[:ulid], clients_obj[ulid+1:]...) // 删除
                        clients = append(clients[:ulid], clients[ulid+1:]...) // 删除
                        clients_obj = append(clients_obj[:upid], clients_obj[upid+1:]...) // 删除
                        clients = append(clients[:upid], clients[upid+1:]...) // 删除
                        fmt.Printf("%+v\n", clients_obj)
                }
                fmt.Println(timeDate() + string(buffer[0:n]))

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
