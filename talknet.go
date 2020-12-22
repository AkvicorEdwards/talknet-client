package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"talknet-client/def"
	"talknet-client/tcp"
	"time"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Base recover", err)
		}
	}()

	serverIp := flag.String("s", "", "-s server ip")
	register := flag.Bool("r", false, "-r register user")
	u := flag.String("u", "", "-u username")
	p := flag.String("p", "", "-p password")
	flag.Parse()
	if *serverIp == "" {
		*serverIp = def.ServerIP
	}
	if *register && *u != "" && *p != "" {
		Register(fmt.Sprintf("%s%s", *serverIp, def.MainPort), *u, *p)
	}

	go ShutDownListener()

	CheckDir()

	rand.Seed(time.Now().UnixNano())

	var username, password string
	fmt.Print("Enter Username: ")
	_, _ = fmt.Scanf("%s", &username)
	fmt.Print("Enter Password: ")
	_, _ = fmt.Scanf("%s", &password)

	tcp.ConnectServer(fmt.Sprintf("%s%s", *serverIp, def.MainPort), username, password)
}

func Register(address, username, password string) {
	var (
		data = make([]byte, tcp.LengthHeadPackage+10)
		// 数据长度
		n   int
		pkg = tcp.NewPackage()
	)
	//pkg := tcp.NewPackage()
	pkg.SetRequestCode(def.RegisterRequest)
	d, ok := tcp.WrapLoginData(username, password)
	if !ok {
		log.Println("Register Error")
	}
	pkg.SetHeadData(d)
	pkg.SetTime(uint64(time.Now().UnixNano()))
	pkg.SetHeadCheckSum()
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Println("Failed to ResolveTCPAddr:", err.Error())
		return
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Println("Failed to DialTCP:", err.Error())
		return
	}
	err = conn.SetWriteDeadline(time.Now().Add(20 * time.Second))
	if err != nil {
		_ = conn.Close()
		return
	}
	n, err = conn.Write(pkg.Data())
	if err != nil || n != tcp.LengthHeadPackage {
		log.Println("Error: Send connection request")
		_ = conn.Close()
		return
	}
	err = conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	if err != nil {
		_ = conn.Close()
		return
	}
	n, err = conn.Read(data)
	if err != nil || n != tcp.LengthHeadPackage {
		log.Println("Error: Receive connection response")
		_ = conn.Close()
		return
	}
	pkg = tcp.ConvertToPackage(data[:n])
	if !pkg.CheckHeadCheckSum() || pkg.GetRequestCode() != def.RegisterRespond {
		log.Println("Error: Check connection response")
		_ = conn.Close()
		return
	}
	fmt.Println(string(pkg.GetHeadData()))

}

func ShutDownListener() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("ShutDownListener recover", err)
		}
	}()
	down := make(chan os.Signal, 1)
	signal.Notify(down, os.Interrupt, os.Kill)
	<-down
	log.Println("Preparing to close")
	if tcp.Cli != nil {
		p := tcp.NewPackage()
		p.SetRequestCode(def.TerminateTheConnection)
		tcp.Cli.DataSend <- &p
		time.Sleep(1*time.Second)
		tcp.Cli.Termination <- true
	} else {
		os.Exit(0)
	}

}

func CheckDir() {
	if _, err := os.Stat(def.TempDir); err != nil {
		fmt.Println("path not exists ", def.TempDir)
		err := os.MkdirAll(def.TempDir, 0711)

		if err != nil {
			log.Println("Error creating directory")
			log.Println(err)
			os.Exit(0)
		}
	}
}
