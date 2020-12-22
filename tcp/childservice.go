package tcp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"talknet-client/def"
)

// 响应心跳请求
func RespondHeartbeatRequest(p *Package, cli *Connection) {
	p.ClearExceptSeq()
	p.SetRequestCode(def.HeartbeatRespond)
	p.SetACK(p.GetSEQ())
	cli.DataSend <- p
}

// 显示信息
func CheckMessage(p *Package, cli *Connection) {
	uuid, data := UnwrapMessage(p.GetHeadData())
	if p.GetExtendedDataFlag() != 1 {
		if uuid == 0 {
			log.Printf("S:[%s]\n", data)
		} else {
			log.Printf("Receive message from: U[%d] [%s]\n", uuid, data)
		}
	} else {
		log.Println("Prepare long message")
		FileReceivePrepare(data, strings.Split(cli.Connection.RemoteAddr().String(), ":")[0],
			cli.UUID, uuid, p.GetExternalDataCheckSum())
		txt := FileReceiveAndRead(data, p.GetExternalDataCheckSum())
		if uuid == 0 {
			log.Printf("S:[%s]\n", string(txt))
		} else {
			log.Printf("Receive long message from: U[%d] [%s]\n", uuid, string(txt))
		}
		//fmt.Printf("Receive long message from: U[%d] [%s]\n", uuid, string(txt))
		FileMapMutex.Lock()
		delete(FileMap, data)
		FileMapMutex.Unlock()
	}
}

// 显示信息
func CheckGroupMessage(p *Package, cli *Connection) {
	guid, uuid, data := UnwrapGroupMessage(p.GetHeadData())
	if p.GetExtendedDataFlag() != 1 {
		fmt.Printf("Group message from: G[%d] U[%d] [%s]\n", guid, uuid, data)
	} else {
		FileReceivePrepare(data, strings.Split(cli.Connection.RemoteAddr().String(), ":")[0],
			cli.UUID, uuid, p.GetExternalDataCheckSum())
		txt := FileReceiveAndRead(data, p.GetExternalDataCheckSum())
		fmt.Printf("Group long message from: G[%d] U[%d] [%s]\n", guid, uuid, string(txt))
		FileMapMutex.Lock()
		delete(FileMap, data)
		FileMapMutex.Unlock()
	}
}

// 发送长信息所在的文件
func CheckPermitLongMessage(p *Package) {
	_, filename := UnwrapMessage(p.GetHeadData())
	FileSend(fmt.Sprintf("%s%s", def.ServerIP, def.FileReceivePort), def.TempDir, filename, true)
}

// 发送长信息所在的文件
func CheckPermitLongGroupMessage(p *Package) {
	_, _, filename := UnwrapGroupMessage(p.GetHeadData())
	FileSend(fmt.Sprintf("%s%s", def.ServerIP, def.FileReceivePort), def.TempDir, filename, true)
}

// 发送长信息所在的文件
func CheckPermitSendFile(p *Package) {
	_, filename := UnwrapMessage(p.GetHeadData())
	FileSend2(fmt.Sprintf("%s%s", def.ServerIP, def.FileReceivePort), filename, false)
}

// 发送长信息所在的文件
func CheckPermitSendGroupFile(p *Package) {
	_, filename := UnwrapMessage(p.GetHeadData())
	FileSend2(fmt.Sprintf("%s%s", def.ServerIP, def.FileReceivePort), filename, false)
}

// 发送长信息所在的文件
func CheckPermitDownloadGroupFile(p *Package) {
	//log.Println(p.GetHeadData())
	//_, filename := UnwrapMessage(p.GetHeadData())
	// TODO
	FileReceive(fmt.Sprintf("%s%s", def.ServerIP, def.FileSendPort), string(p.GetHeadData()))
}

// 检查好友邀请列表
func CheckListFriendInvitation(p *Package) {
	data := make([]uint32, 0)
	if p.GetExtendedDataFlag() == 0 {
		_, ms := UnwrapMessage(p.GetHeadData())
		err := json.Unmarshal([]byte(ms), &data)
		if err != nil {
			log.Println(string(p.GetHeadData()), p.GetHeadData(), err)
			return
		}
	} else {
		// 准备接受文件
		filename := string(p.GetHeadData())

		FileReceive(fmt.Sprintf("%s%s", def.ServerIP, def.FileSendPort), filename)
		txt, err := ioutil.ReadFile(def.TempDir + filename)
		_ = os.Remove(def.TempDir + filename)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(string(txt))
		if CRC32(txt) != p.GetExternalDataCheckSum() {
			fmt.Printf("CRC32 Need:[%d] Get:[%d]\n", p.GetExternalDataCheckSum(), CRC32(txt))
			return
		}
		err = json.Unmarshal(txt, &data)
		if err != nil {
			log.Println(err)
			return
		}
	}
	fmt.Println("Unaccepted Friend Invitation")
	for _, v := range data {
		fmt.Printf("%d ", v)
	}
	fmt.Println("")
	fmt.Println("Input 'uuid' to accept invitation:")
}

func CheckListFriend(p *Package) {
	data := make([]uint32, 0)
	if p.GetExtendedDataFlag() == 0 {
		_, ms := UnwrapMessage(p.GetHeadData())
		err := json.Unmarshal([]byte(ms), &data)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		// 准备接受文件
		filename := string(p.GetHeadData())

		FileReceive(fmt.Sprintf("%s%s", def.ServerIP, def.FileSendPort), filename)
		txt, err := ioutil.ReadFile(def.TempDir + filename)
		_ = os.Remove(def.TempDir + filename)
		if err != nil {
			log.Println(err)
			return
		}
		if CRC32(txt) != p.GetExternalDataCheckSum() {
			log.Printf("CRC32 Need:[%d] Get:[%d]\n", p.GetExternalDataCheckSum(), CRC32(txt))
			return
		}
		err = json.Unmarshal(txt, &data)
		if err != nil {
			log.Println(err)
			return
		}
	}
	fmt.Println("Friends")
	for _, v := range data {
		fmt.Printf("%d ", v)
	}
	fmt.Println("")
	fmt.Println("Input 'uuid' to chat:")
}

func CheckListGroup(p *Package) {
	data := make([]uint32, 0)
	if p.GetExtendedDataFlag() == 0 {
		_, ms := UnwrapMessage(p.GetHeadData())
		err := json.Unmarshal([]byte(ms), &data)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		// 准备接受文件
		filename := string(p.GetHeadData())

		FileReceive(fmt.Sprintf("%s%s", def.ServerIP, def.FileSendPort), filename)
		txt, err := ioutil.ReadFile(def.TempDir + filename)
		_ = os.Remove(def.TempDir + filename)
		if err != nil {
			log.Println(err)
			return
		}
		if CRC32(txt) != p.GetExternalDataCheckSum() {
			log.Printf("CRC32 Need:[%d] Get:[%d]\n", p.GetExternalDataCheckSum(), CRC32(txt))
			return
		}
		err = json.Unmarshal(txt, &data)
		if err != nil {
			log.Println(err)
			return
		}
	}
	fmt.Println("Groups")
	for _, v := range data {
		fmt.Printf("%d ", v)
	}
	fmt.Println("")
	fmt.Println("Input 'guid' to manage:")
}

func CheckListGroupMember(p *Package) {
	data := make([]uint32, 0)
	if p.GetExtendedDataFlag() == 0 {
		_, ms := UnwrapMessage(p.GetHeadData())
		err := json.Unmarshal([]byte(ms), &data)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		// 准备接受文件
		filename := string(p.GetHeadData())

		FileReceive(fmt.Sprintf("%s%s", def.ServerIP, def.FileSendPort), filename)
		txt, err := ioutil.ReadFile(def.TempDir + filename)
		_ = os.Remove(def.TempDir + filename)
		if err != nil {
			log.Println(err)
			return
		}
		if CRC32(txt) != p.GetExternalDataCheckSum() {
			log.Printf("CRC32 Need:[%d] Get:[%d]\n", p.GetExternalDataCheckSum(), CRC32(txt))
			return
		}
		err = json.Unmarshal(txt, &data)
		if err != nil {
			log.Println(err)
			return
		}
	}
	fmt.Println("Groups Member")
	for _, v := range data {
		fmt.Printf("%d ", v)
	}
	fmt.Println("")
	fmt.Println("Input 'uuid' to manage:")
}

func CheckListGroupFile(p *Package) {
	data := make([]def.Files, 0)
	if p.GetExtendedDataFlag() == 0 {
		_, ms := UnwrapMessage(p.GetHeadData())
		//log.Println(p.GetHeadData())
		err := json.Unmarshal([]byte(ms), &data)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		// 准备接受文件
		filename := string(p.GetHeadData())

		FileReceive(fmt.Sprintf("%s%s", def.ServerIP, def.FileSendPort), filename)
		txt, err := ioutil.ReadFile(def.TempDir + filename)
		_ = os.Remove(def.TempDir + filename)
		if err != nil {
			log.Println(err)
			return
		}
		if CRC32(txt) != p.GetExternalDataCheckSum() {
			log.Printf("CRC32 Need:[%d] Get:[%d]\n", p.GetExternalDataCheckSum(), CRC32(txt))
			return
		}
		err = json.Unmarshal(txt, &data)
		if err != nil {
			log.Println(err)
			return
		}
	}
	fmt.Println("Groups Files")
	for _, v := range data {
		fmt.Printf("FUID: [%d] UUID: [%d] Name: [%s]\n", v.Fuid, v.Uuid, v.RealName)
	}
	fmt.Println("")
	fmt.Println("Input 'fuid' to download:")
}

func CheckListGroupAdmin(p *Package) {
	data := make([]uint32, 0)
	if p.GetExtendedDataFlag() == 0 {
		_, ms := UnwrapMessage(p.GetHeadData())
		err := json.Unmarshal([]byte(ms), &data)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		// 准备接受文件
		filename := string(p.GetHeadData())

		FileReceive(fmt.Sprintf("%s%s", def.ServerIP, def.FileSendPort), filename)
		txt, err := ioutil.ReadFile(def.TempDir + filename)
		_ = os.Remove(def.TempDir + filename)
		if err != nil {
			log.Println(err)
			return
		}
		if CRC32(txt) != p.GetExternalDataCheckSum() {
			log.Printf("CRC32 Need:[%d] Get:[%d]\n", p.GetExternalDataCheckSum(), CRC32(txt))
			return
		}
		err = json.Unmarshal(txt, &data)
		if err != nil {
			log.Println(err)
			return
		}
	}
	fmt.Println("Groups Admin")
	for _, v := range data {
		fmt.Printf("%d ", v)
	}
	fmt.Println("")
	fmt.Println("Input 'uuid' to revoke:")
}

func CheckListJoinGroup(p *Package) {
	data := make([]uint32, 0)
	if p.GetExtendedDataFlag() == 0 {
		_, ms := UnwrapMessage(p.GetHeadData())
		err := json.Unmarshal([]byte(ms), &data)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		// 准备接受文件
		filename := string(p.GetHeadData())

		FileReceive(fmt.Sprintf("%s%s", def.ServerIP, def.FileSendPort), filename)
		txt, err := ioutil.ReadFile(def.TempDir + filename)
		_ = os.Remove(def.TempDir + filename)
		if err != nil {
			log.Println(err)
			return
		}
		if CRC32(txt) != p.GetExternalDataCheckSum() {
			log.Printf("CRC32 Need:[%d] Get:[%d]\n", p.GetExternalDataCheckSum(), CRC32(txt))
			return
		}
		err = json.Unmarshal(txt, &data)
		if err != nil {
			log.Println(err)
			return
		}
	}
	fmt.Println("Group application")
	for _, v := range data {
		fmt.Printf("%d ", v)
	}
	fmt.Println("")
	fmt.Println("Input 'uuid' to accept:")
}

// 自适应接受数据
func AutoAdaptReceiveData(p *Package) []byte {
	if p.GetExtendedDataFlag() == 0 {
		// 短数据，直接提取并返回
		return p.GetHeadData()
	} else {
		// 长数据，向服务器请求接收文件，返回文件内容并删除接受的文件
		filename := string(p.GetHeadData())
		return FileReceiveAndRead(filename, p.GetExternalDataCheckSum())
	}
}

func FileReceivePrepare(filename, ip string, from, to, checkSum uint32) *FileMapInfo {
	// 将信息打包以filename为key存储在 FileMap 中
	info := &FileMapInfo{
		Ip:       ip,
		From:     from,
		To:       to,
		CheckSum: checkSum,
		Mutex:    sync.Mutex{},
	}

	FileMapMutex.Lock()
	FileMap[filename] = info
	FileMapMutex.Unlock()

	return info
}

func FileSendPrepare(filename, path string, from, to, checkSum uint32) *FileMapInfo {
	// 将信息打包以filename为key存储在 FileMap 中
	info := &FileMapInfo{
		Path:     path,
		From:     from,
		To:       to,
		CheckSum: checkSum,
		Mutex:    sync.Mutex{},
	}

	FileMapMutex.Lock()
	FileMap[filename] = info
	FileMapMutex.Unlock()

	return info
}

// 接收文件，返回文件数据，删除文件
func FileReceiveAndRead(filename string, checkSum uint32) []byte {
	FileReceive(fmt.Sprintf("%s%s", def.ServerIP, def.FileSendPort), filename)
	txt, err := ioutil.ReadFile(def.TempDir + filename)
	_ = os.Remove(def.TempDir + filename)
	if err != nil {
		log.Println(err)
		return []byte{}
	}
	if CRC32(txt) != checkSum {
		log.Printf("CRC32 Need:[%d] Get:[%d]\n", checkSum, CRC32(txt))
		return []byte{}
	}
	return txt
}

func ReceiveFile(p *Package) {
	uuid, filename := UnwrapMessage(p.GetHeadData())
	log.Printf("Preparing receive file [%s] from [%d]", filename, uuid)
	FileReceive(fmt.Sprintf("%s%s", def.ServerIP, def.FileSendPort), filename)

	_, _, hashVal := CalculateFileHashValue(def.TempDir + filename)

	if hashVal != p.GetExternalDataCheckSum() {
		log.Printf("CRC32 Need:[%d] Get:[%d]\n", p.GetExternalDataCheckSum(), hashVal)
	} else {
		log.Printf("Receive file [%s] from [%d] CRC32[%d]\n", filename, uuid, hashVal)
	}

}
