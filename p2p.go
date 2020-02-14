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
    "log"
)

// conflag
var conf struct {
    isServer bool
    isClient bool
    LocalAddress string
    LocalKey string
    PeerKey string
    RemoteAddress string
    RemoteCryptKey string
    PeerCryptKey string
    isGenKey bool
}

func main() {
    flag := conflag.New()

    flag.BoolVar(&conf.isServer,"s",false,"Whether it is a server")
    flag.BoolVar(&conf.isClient,"c",false,"Whether it is a client")
    flag.StringVar(&conf.LocalAddress,"l","0.0.0.0:9000","Local Listen Address")
    flag.StringVar(&conf.LocalKey,"lk","99998","Local Key")
    flag.StringVar(&conf.PeerKey,"pk","99999","Peer Key")
    flag.StringVar(&conf.RemoteAddress,"r","127.0.0.1:9000","Remote Server Address")
    flag.StringVar(&conf.RemoteCryptKey,"sk","defaultremotekey","Remote Server Encryption key")
    flag.StringVar(&conf.PeerCryptKey,"ck","defaultpeerkey","Peer Encryption key")
    flag.BoolVar(&conf.isGenKey,"g",false,"Whether generate keys")

    flag.Parse()

    //log.Printf("%+v\n",conf)

    if conf.isServer && conf.isClient {
        log.Println("It is not allowed to run server and client at the same time!")
    }else if conf.isServer && !conf.isClient {
        runAsServer(conf.LocalAddress)
    }else if !conf.isServer && conf.isClient {
        runAsClient(conf.RemoteAddress, conf.LocalKey, conf.PeerKey)
    }else if conf.isGenKey {
        patchKeysetHandle(getBaseDirPath() + "/peer.dc")
    }else {
        log.Println("No operating mode specified!")
    }
}
// runAsClient
// 作为客户端运行
func runAsClient(CONNECT, LOCALKEY, PeerKey string) {
    patchKeysetHandleFromJSON(getBaseDirPath() + "/server.dc")//Load server-side encryption key

    srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 1999}
    s, err := net.ResolveUDPAddr("udp4", CONNECT)
    c, err := net.DialUDP("udp4", srcAddr, s)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("The client was started.\n")

    paired,connected := false, false
    var peer net.UDPAddr

    for {
        data := LOCALKEY + "|" + PeerKey
        PeerKeyLength := utf8.RuneCountInString(data) - 1
        if PeerKeyLength > 32 && PeerKeyLength < 16 {
            log.Println("PeerKey not allowed to be too long")
            return
        }
        _, err = c.Write(cryptOfAEAD([]byte(data), []byte(conf.RemoteCryptKey), false))

        if err != nil {
            log.Fatal(err)
        }

        buffer := make([]byte, 1024)
        n, _, err := c.ReadFromUDP(buffer)
        if err != nil {
            log.Fatal(err)
        }

        if(!paired){
            paired = true
            peer = parseAddr(string(cryptOfAEAD(buffer[0:n], []byte(conf.RemoteCryptKey), true)))
            log.Printf("Obtained the peer address：%s\n", string(cryptOfAEAD(buffer[0:n], []byte(conf.RemoteCryptKey), true)))
            c.Close()
        }
        
        break
    }

    patchKeysetHandleFromJSON(getBaseDirPath() + "/peer.dc")//Load peer encryption key
    log.Println("Waiting before punching")
    if(paired && !connected){
        log.Println("Punching starts")
        conn, err := net.DialUDP("udp", srcAddr, &peer)
        if err != nil {
            log.Fatal(err)
        }
        connected = true

        defer conn.Close()

        if _, err = conn.Write(cryptOfAEAD(genMsg(conf.LocalKey, "peer", timestampNano()), []byte(conf.PeerCryptKey), false)); err != nil {
            log.Println("send handshake:", err)
        }
        // 心跳包
        go func() {
            for {
                time.Sleep(3 * time.Second)
                data := cryptOfAEAD(genMsg(conf.LocalKey, "ping", timestampNano()), []byte(conf.PeerCryptKey), false)
                if _, err = conn.Write(data); err != nil {
                    log.Println("ping fail", err)
                }
            }
        }()
        go func() {
            for {
                //time.Sleep(10 * time.Second)
                reader := bufio.NewReader(os.Stdin)
                text, _ := reader.ReadString('\n')
                data := cryptOfAEAD(genMsg(conf.LocalKey,strings.Replace(text, "\n", "", -1), timestampNano()), []byte(conf.PeerCryptKey), false)//去除回车符
                if _, err = conn.Write(data); err != nil {
                    log.Println("send msg fail", err)
                }
            }
        }()
        for {
            data := make([]byte, 1024)
            n, _, err := conn.ReadFromUDP(data)
            if err != nil {
                log.Printf("error during read: %s\n", err)
            } else {
                dataString := cryptOfAEAD(data[:n], []byte(conf.PeerCryptKey), true)
                msg := parserMsg(dataString)
                if stringCompare([]byte(msg.Body), []byte("ping")) != 0 {
                    log.Printf("%s\t%dms\n", msg.Body, timestampNano() - msg.TimeStamp)
                }
            }
        }
    }

}
// runAsServer
// 作为服务端运行
func runAsServer(PORT string) {
    patchKeysetHandleFromJSON(getBaseDirPath() + "/server.dc")//Load server-side encryption key

    s, err := net.ResolveUDPAddr("udp4", PORT)
    if err != nil {
        log.Fatal(err)
    }

    connection, err := net.ListenUDP("udp4", s)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("The server was started.\n")

    defer connection.Close()
    buffer := make([]byte, 1024)

    clients := make([]net.UDPAddr, 0, 20)//Up to 20 clients at the same time
    clients_obj := make([]map[string]string, 0, 20)//Up to 20 clients at the same time

    for {
        n, addr, err := connection.ReadFromUDP(buffer)
        if err != nil {
            log.Fatal(err)
        }

        if !checkCryptOfAEAD(buffer[0:n], []byte(conf.RemoteCryptKey), true) {
            log.Println("Pre-encryption and decryption failed!")
            continue
        }
        dataString := cryptOfAEADString(buffer[0:n], []byte(conf.RemoteCryptKey), true)
        if !strings.Contains(dataString, "|") {
            continue
        }
        data := strings.Split(dataString, "|")

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
            //log.Printf("%+v\n", clients)
            client := make(map[string]string)
            client["localKey"] = data[0]
            client["localId"] = fmt.Sprintf("%d", ulid)
            client["PeerKey"] = data[1]
            client["pairedId"] = fmt.Sprintf("%d", upid)
            client["clientAddr"] = addr.String()
            clients_obj = append(clients_obj, client)
            client["localId"] = fmt.Sprintf("%d", len(clients_obj)-1)
            ulid = len(clients_obj)-1
            //log.Printf("%+v\n", clients_obj)
        }

        log.Printf(timeDate() + "-> (%d)LocalKey: %s | (%d)PeerKey: %s\n", ulid, data[0], upid, data[1])

        if ulid != -1 && upid != -1 {
            connection.WriteToUDP(cryptOfAEAD([]byte(clients[ulid].String()), []byte(conf.RemoteCryptKey), false), &clients[upid])
            connection.WriteToUDP(cryptOfAEAD([]byte(clients[upid].String()), []byte(conf.RemoteCryptKey), false), &clients[ulid])
            // 待加入动态删除
            clients_obj = append(clients_obj[:ulid], clients_obj[ulid+1:]...) // 删除
            clients = append(clients[:ulid], clients[ulid+1:]...) // 删除
            clients_obj = append(clients_obj[:upid], clients_obj[upid+1:]...) // 删除
            clients = append(clients[:upid], clients[upid+1:]...) // 删除
            log.Printf("%+v\n", clients_obj)
        }
        log.Println(string(cryptOfAEAD(buffer[0:n], []byte(conf.RemoteCryptKey), true)))

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
// timestampNano
// 返回时间戳，毫秒
func timestampNano() int64 {
    return time.Now().UnixNano() / 1e6
}