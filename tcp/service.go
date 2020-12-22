package tcp

import (
	"fmt"
	"log"
	"net"
	"os"
	"talknet-client/def"
	"time"
)

var Cli *Connection

// 连接服务器并登录
func ConnectServer(address string, username, password string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("ConnectServer recover", err)
		}
	}()
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

	var (
		data = make([]byte, LengthHeadPackage+10)
		// 数据长度
		n   int
		pkg = NewPackage()
	)

	// 发送心跳请求
	pkg.SetRequestCode(def.HeartbeatRequest)
	pkg.SetSEQ(1)
	pkg.SetHeadCheckSum()
	err = conn.SetWriteDeadline(time.Now().Add(20 * time.Second))
	if err != nil {
		_ = conn.Close()
		return
	}
	n, err = conn.Write(pkg.Data())
	if err != nil || n != LengthHeadPackage {
		log.Println("Error: Send connection request")
		_ = conn.Close()
		return
	}

	// 接收心跳回应
	err = conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	if err != nil {
		_ = conn.Close()
		return
	}
	n, err = conn.Read(data)
	if err != nil || n != LengthHeadPackage {
		log.Println("Error: Receive connection response")
		_ = conn.Close()
		return
	}
	pkg = ConvertToPackage(data[:n])
	if !pkg.CheckHeadCheckSum() || pkg.GetACK() != 1 ||
		pkg.GetRequestCode() != def.HeartbeatRespond {
		log.Println("Error: Check connection response")
		_ = conn.Close()
		return
	}

	// 建立连接成功，发送登录数据
	pkg.ClearExceptSeq()
	pkg.SetRequestCode(def.Login)
	pkg.SetSEQ(2)
	userData, ok := WrapLoginData(username, password)
	if !ok {
		log.Println("Error: Illegal login data")
		_ = conn.Close()
		return
	}
	pkg.SetHeadData(userData)
	pkg.SetHeadCheckSum()
	err = conn.SetWriteDeadline(time.Now().Add(20 * time.Second))
	if err != nil {
		_ = conn.Close()
		return
	}
	n, err = conn.Write(pkg.Data())
	if err != nil || n != LengthHeadPackage {
		log.Println("Error: Send login data")
		_ = conn.Close()
		return
	}

	// 接收登录回应
	err = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		_ = conn.Close()
		return
	}
	n, err = conn.Read(data)
	if err != nil || n != LengthHeadPackage {
		log.Println("Error: Receive login response")
		_ = conn.Close()
		return
	}
	pkg = ConvertToPackage(data[:n])
	if !pkg.CheckHeadCheckSum() || pkg.GetACK() != 2 {
		log.Println("Error: Check login response")
		_ = conn.Close()
		return
	}

	if pkg.GetRequestCode() == def.LoginSuccessful {
		fmt.Println("Login Successful")
		_id, _uname, _niname := UnwrapUserInfo(pkg.GetHeadData())
		fmt.Printf("UUID:[%d] Username:[%s] Nickname:[%s]\n", _id, _uname, _niname)
	} else {
		fmt.Println("Login Failure.", string(pkg.GetHeadData()))
		return
	}

	uuid, username, nickname := UnwrapUserInfo(pkg.GetHeadData())
	cli := NewConnection(uuid, username, nickname, conn)
	Cli = cli
	go Terminator(cli)
	go Receiver(cli)
	go Sender(cli)
	go Heartbeat(cli)
	go TerminalConsole(cli)
	select {}
}

// 信息发送服务
// 监听 Connection.DataSend 中的信息并发送
func Sender(cli *Connection) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Sender recover", err)
		}
	}()

	for {
		select {
		case <-cli.WorkerReq.Sender:
			cli.WorkerRes.Sender <- true
			return
		case d := <-cli.DataSend:
			// 设置SEQ
			cli.SEQMutex.Lock()
			d.SetSEQ(cli.SEQ)
			cli.SEQ++
			cli.SEQMutex.Unlock()
			// 设置时间戳
			d.SetTime(uint64(time.Now().UnixNano()))
			d.SetHeadCheckSum()

			PrintPackage(d, false, false)
			_ = cli.Connection.SetWriteDeadline(time.Now().Add(20 * time.Second))
			n, err := cli.Connection.Write(d.Data())
			if err != nil {
				log.Printf("UUID:[%v] Error sending data:"+
					" failed to send. [%s]\n", cli.UUID, err.Error())
				continue
			}
			if n != LengthHeadPackage {
				log.Printf("UUID:[%v] Error sending data:"+
					" the length of the sent data does not match. "+
					"%d bytes sent. actual length %d bytes\n",
					cli.UUID, n, LengthHeadPackage)
				continue
			}
			cli.ResetHeartbeat <- true
		}
	}
}

// 接收并处理信息
func Receiver(cli *Connection) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Receiver recover", err)
		}
	}()

	dataT := make([]byte, LengthHeadPackage+10)
	down := false
	go func() {
		for {
			select {
			case <-cli.WorkerReq.Receiver:
				down = true
				return
			}
		}
	}()

	for {
		if down {
			cli.WorkerRes.Receiver <- true
			return
		}

		_ = cli.Connection.SetReadDeadline(time.Now().Add(20 * time.Second))
		n, err := cli.Connection.Read(dataT)

		// 发送心跳数据
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		if n != LengthHeadPackage {
			log.Println(n)
			continue
		}

		// 收到数据，重置Heartbeat循环
		cli.ResetHeartbeat <- true

		data := ConvertToPackage(dataT[:n])
		if !data.CheckHeadCheckSum() || time.Now().UnixNano()-int64(data.GetTime()) >= int64(10 * time.Second) {
			fmt.Println("******** !!!Broken!!! *******")
			PrintPackage(&data, true, true)
			continue
		}

		PrintPackage(&data, false, true)
		ProcessPackage(&data, cli)
	}
}

func ProcessPackage(p *Package, cli *Connection) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("ProcessPackage recover", err)
		}
	}()

	switch p.GetRequestCode() {
	case def.HeartbeatRequest:
		RespondHeartbeatRequest(p, cli)
	case def.Message:
		CheckMessage(p, cli)
	case def.PermitLongMessage:
		CheckPermitLongMessage(p)
	case def.ListFriendInvitation:
		CheckListFriendInvitation(p)
	case def.ListFriend:
		CheckListFriend(p)
	case def.ListGroup:
		CheckListGroup(p)
	case def.ListGroupMember:
		CheckListGroupMember(p)
	case def.ListJoinGroup:
		CheckListJoinGroup(p)
	case def.ListGroupAdmin:
		CheckListGroupAdmin(p)
	case def.TerminateTheConnection:
		cli.Termination <- true
	case def.GroupMessage:
		CheckGroupMessage(p, cli)
	case def.PermitLongGroupMessage:
		CheckPermitLongGroupMessage(p)
	case def.PermitSendFile:
		CheckPermitSendFile(p)
	case def.SendFile:
		ReceiveFile(p)
	case def.PermitSendGroupFile:
		CheckPermitSendGroupFile(p)
	case def.ListGroupFile:
		CheckListGroupFile(p)
	case def.PermitDownloadGroupFile:
		CheckPermitDownloadGroupFile(p)
	default:
		fmt.Println("******* !!!Rubbish!!! *******")
		PrintPackage(p, true, true)
	}
}

// 心跳监测
// 若1分钟内没有任何数据发送或接收，则关闭程序
func Heartbeat(cli *Connection) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Heartbeat recover", err)
		}
	}()
	for {
		select {
		case <-cli.WorkerReq.Heartbeat:
			cli.WorkerRes.Heartbeat <- true
			return
		case <-cli.ResetHeartbeat:
		case <-time.After(1 * time.Minute):
			cli.Termination <- true
		}
	}
}

// 退出程序
func Terminator(cli *Connection) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Terminator recover", err)
		}
	}()

	select {
	case <-cli.Termination:
		log.Printf("UUID:[%v] Kill Signal Generated\n", cli.UUID)
		log.Println("Remove temp files")
		FileMapMutex.Lock()
		for k, v := range FileMap {
			v.Mutex.Lock()
			err := os.Remove(def.TempDir + k)
			if err != nil {
				log.Println("Error: Remove file", k, err)
			}
		}
		os.Exit(0)
	}
}
