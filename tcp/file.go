package tcp

import (
	"io"
	"log"
	"net"
	"os"
	"sync"
	"talknet-client/def"
	"time"
)

type FileMapInfo struct {
	Path     string
	Ip       string
	From     uint32
	To       uint32
	CheckSum uint32
	Mutex    sync.Mutex
}

var FileMap = make(map[string]*FileMapInfo)
var FileMapMutex = sync.Mutex{}

// 1. 写 文件名
// 3. 读 文件
// 向服务器发送请求，接收指定文件
func FileReceive(address, filename string) {
	log.Printf("FileReceive [%s]\n", filename)
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
	defer func() { _ = conn.Close() }()

	_, err = conn.Write([]byte(filename))
	if err != nil {
		log.Println("Write filename", filename, err)
		return
	}

	fs, err := os.Create(def.TempDir + filename)
	if err != nil {
		log.Println("Create")
		return
	}
	defer func() {
		err = fs.Close()
		if err != nil {
			log.Println("Receive Close", err)
		}
	}()

	err = conn.SetReadDeadline(time.Now().Add(3 * time.Hour))
	if err != nil {
		_ = conn.Close()
		return
	}
	err = conn.SetDeadline(time.Now().Add(3 * time.Hour))
	if err != nil {
		_ = conn.Close()
		return
	}

	buf := make([]byte, 1024*10)
	for {
		n, err := conn.Read(buf)
		if err != nil || n == 0 {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}
		_, _ = fs.Write(buf[:n])
	}

}

// 1. 写 文件名
// 2. 读 ok
// 3. 写 文件
func FileSend2(address string, filename string, delete bool) {
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

	defer func() { _ = conn.Close() }()

	f, ok := FileMap[filename]
	if !ok {
		// TODO
		log.Println("File do not exit")
		return
	}

	_, _ = conn.Write([]byte(filename))
	//log.Println("Send file", filename)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("conn.Read err =", err)
		return
	}

	if string(buf[:n]) != "ok" {
		log.Println("not ok")
		return
	}
	log.Println("Sending file")

	fs, err := os.Open(f.Path + filename)
	if err != nil {
		log.Println("os.Open err =", err)
		return
	}
	defer func() {
		err = fs.Close()
		if err != nil {
			log.Println("Close file with error:", err)
		}
		//log.Println("Send file", def.TempDir+filename)
		if delete {
			err = os.Remove(f.Path + filename)
			if err != nil {
				log.Println("Remove file with error:", err)
			}
		}
		//log.Println("Close file\nRemove File")
	}()


	err = conn.SetWriteDeadline(time.Now().Add(3 * time.Hour))
	if err != nil {
		_ = conn.Close()
		return
	}
	err = conn.SetDeadline(time.Now().Add(3 * time.Hour))
	if err != nil {
		_ = conn.Close()
		return
	}

	buf = make([]byte, 1024*10)
	for {
		n, err = fs.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}
		_, _ = conn.Write(buf[:n])
	}

}

// 1. 写 文件名
// 2. 读 ok
// 3. 写 文件
func FileSend(address string, dir, filename string, delete bool) {
	log.Printf("FileSend [%s]", filename)
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

	defer func() { _ = conn.Close() }()

	_, _ = conn.Write([]byte(filename))
	//log.Println("Send file", filename)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("conn.Read err =", err)
		return
	}

	if string(buf[:n]) != "ok" {
		log.Println("not ok")
		return
	}

	fs, err := os.Open(dir + filename)
	if err != nil {
		log.Println("os.Open err =", err)
		return
	}

	defer func() {
		err = fs.Close()
		if err != nil {
			log.Println("Close file with error:", err)
		}
		//log.Println("Send file", def.TempDir+filename)
		if delete {
			err = os.Remove(dir + filename)
			if err != nil {
				log.Println("Remove file with error:", err)
			}
		}
		//log.Println("Close file\nRemove File")
	}()

	err = conn.SetWriteDeadline(time.Now().Add(3 * time.Hour))
	if err != nil {
		_ = conn.Close()
		return
	}

	buf = make([]byte, 1024*10)
	for {
		n, err = fs.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}
		_, _ = conn.Write(buf[:n])
	}

}
