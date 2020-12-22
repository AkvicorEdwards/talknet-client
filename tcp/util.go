package tcp

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"talknet-client/def"
	"time"
)

func UInt16ToBytes(i uint16) []byte {
	var buf = make([]byte, 2)
	binary.BigEndian.PutUint16(buf, i)
	return buf
}

func BytesToUInt16(buf []byte) uint16 {
	return binary.BigEndian.Uint16(buf)
}

func UInt32ToBytes(i uint32) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, i)
	return buf
}

func BytesToUInt32(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf)
}

func UInt64ToBytes(i uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, i)
	return buf
}

func BytesToUInt64(buf []byte) uint64 {
	return binary.BigEndian.Uint64(buf)
}

func CRC32(str []byte) uint32 {
	return crc32.ChecksumIEEE(str)
}

// 将账号密码打包
// username length limit:  def.LoginDataLengthUsername bytes
// password length limit: def.LoginDataLengthPassword bytes
func WrapLoginData(username, password string) ([]byte, bool) {
	if len([]byte(username)) > def.LoginDataLengthUsername || len([]byte(password)) > def.LoginDataLengthPassword {
		return []byte{}, false
	}
	data := make([]byte, LengthHeadData)
	user := []byte(username)
	pass := []byte(password)
	data[def.LoginDataOffsetUsernameLength] = byte(len(user))
	data[def.LoginDataOffsetPasswordLength] = byte(len(pass))
	for k, v := range user {
		data[def.LoginDataOffsetUsername+k] = v
	}
	for k, v := range pass {
		data[def.LoginDataOffsetPassword+k] = v
	}
	return data, true
}

// 解包用户信息
func UnwrapUserInfo(data []byte) (uuid uint32, username, nickname string) {
	if len(data) != LengthHeadData {
		return 0, "", ""
	}
	return BytesToUInt32(data[def.UserInfoOffsetUUID:def.UserInfoOffsetUUID+def.UserInfoLengthUUID]),
		string(data[def.UserInfoOffsetUsername:def.UserInfoOffsetUsername+data[def.UserInfoOffsetUsernameLength]]),
		string(data[def.UserInfoOffsetNickname:def.UserInfoOffsetNickname+data[def.UserInfoOffsetNicknameLength]])
}

// 打包短信息
func WrapMessage(uuid uint32, message string) ([]byte, bool)  {
	if len([]byte(message)) >= def.MessageLengthMessage {
		return []byte{}, false
	}
	data := make([]byte, LengthHeadData)
	user := UInt32ToBytes(uuid)
	mess := []byte(message)
	data[def.MessageOffsetMessageLength] = byte(len(mess))
	for k, v := range user {
		data[def.MessageOffsetUUID+k] = v
	}
	for k, v := range mess {
		data[def.MessageOffsetMessage+k] = v
	}
	return data, true
}

// 打包短信息
func WrapGroupMessage(guid, uuid uint32, message string) ([]byte, bool)  {
	if len([]byte(message)) >= def.MessageLengthGroupMessage {
		return []byte{}, false
	}
	data := make([]byte, LengthHeadData)
	guidd := UInt32ToBytes(guid)
	uuidd := UInt32ToBytes(uuid)
	mess := []byte(message)
	data[def.MessageOffsetGroupMessageLength] = byte(len(mess))
	for k, v := range guidd {
		data[def.MessageOffsetGroupGUID+k] = v
	}
	for k, v := range uuidd {
		data[def.MessageOffsetGroupUUID+k] = v
	}
	for k, v := range mess {
		data[def.MessageOffsetGroupMessage+k] = v
	}

	return data, true
}

func WrapGuidUuid(guid, uuid uint32) []byte {
	data := UInt32ToBytes(guid)
	data = append(data, UInt32ToBytes(uuid)...)
	return data
}

func UnwrapGuidUuid(data []byte) (guid, uuid uint32) {
	guid = BytesToUInt32(data[:4])
	uuid = BytesToUInt32(data[4:8])
	return
}

func RandomFilename(uuid uint32) string {
	return fmt.Sprintf("%d_%d_%d", uuid, time.Now().UnixNano(), rand.Int63())
}

func WrapLongMessage(uuid uint32, message string) (data []byte, filename string) {
	filename = RandomFilename(uuid)
	err := ioutil.WriteFile(def.TempDir+filename, []byte(message), 0644)
	if err != nil {
		return []byte{}, ""
	}
	data, ok := WrapMessage(uuid, filename)
	if !ok {
		return []byte{}, ""
	}
	return data, filename
}

func WrapLongGroupMessage(guid, uuid uint32, message string) (data []byte, filename string) {
	filename = RandomFilename(uuid)
	err := ioutil.WriteFile(def.TempDir+filename, []byte(message), 0644)
	if err != nil {
		return []byte{}, ""
	}
	data, ok := WrapGroupMessage(guid, uuid, filename)
	if !ok {
		return []byte{}, ""
	}
	return data, filename
}

func UnwrapMessage(data []byte) (uuid uint32, message string) {
	if len(data) != LengthHeadData {
		return 0, ""
	}
	return BytesToUInt32(data[def.MessageOffsetUUID:def.MessageOffsetUUID+def.MessageLengthUUID]),
		string(data[def.MessageOffsetMessage:def.MessageOffsetMessage+data[def.MessageOffsetMessageLength]])
}

func UnwrapGroupMessage(data []byte) (guid, uuid uint32, message string) {
	if len(data) != LengthHeadData {
		return 0, 0, ""
	}
	return BytesToUInt32(data[def.MessageOffsetGroupGUID:def.MessageOffsetGroupGUID+def.MessageLengthGroupGUID]),
		BytesToUInt32(data[def.MessageOffsetGroupUUID:def.MessageOffsetGroupUUID+def.MessageLengthGroupUUID]),
		string(data[def.MessageOffsetGroupMessage:def.MessageOffsetGroupMessage+data[def.MessageOffsetGroupMessageLength]])
}

func ReWrapMessage(from uint32, data Package) Package  {
	ori := data.GetHeadData()
	user := UInt32ToBytes(from)
	for k, v := range user {
		ori[def.MessageOffsetUUID+k] = v
	}
	data.SetHeadData(ori)
	return data
}

func CalculateFileHashValue(path string) (filename, filePath string, hash uint32) {
	val := uint32(0)
	fs, err := os.Open(path)
	if err != nil {
		log.Println("os.Open err =", err)
		return "", "", 0
	}
	defer func() {_=fs.Close()}()
	info, err := fs.Stat()
	if err != nil {
		return "", "", 0
	}
	filename = info.Name()
	filePath = path[:len(path)-len(filename)]
	buf := make([]byte, 1024*10)
	var n int
	for {
		n, err = fs.Read(buf)
		if err != nil {
			break
		}
		val = crc32.Update(val, crc32.IEEETable, buf[:n])
	}
	return filename, filePath, val
}

func PrintPackage(p *Package, per bool, rec bool) {
	if !def.DisplayPackageInfo && !per {
		return
	}
	if rec {
		fmt.Println("*************REC*************")
	} else {
		fmt.Println("*************SED*************")
	}
	fmt.Println("-----------Package-----------")
	fmt.Println("TIME:", p.GetTime(), time.Unix(0, int64(p.GetTime())).Format("2006-01-02 15:04:05"))
	fmt.Println("CODE:", p.GetRequestCode())
	fmt.Println("SEQ :", p.GetSEQ())
	fmt.Println("ACK :", p.GetACK())
	fmt.Println("FLAG:", p.GetExtendedDataFlag())
	fmt.Println("DATA:", p.Data())
	fmt.Println("-----------------------------")
}


func InfoPrintf(format string, v ...interface{}) {
	if !def.DisplayInfo {
		return
	}
	log.Printf(format, v...)
}

func InfoPrintln(v ...interface{}) {
	if !def.DisplayInfo {
		return
	}
	log.Println(v...)
}
