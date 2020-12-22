package tcp

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"talknet-client/def"
)

// 终端操作
func TerminalConsole(cli *Connection) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("TerminalMessageSender recover", err)
		}
	}()
	var inputReader *bufio.Reader
	var input string
	var uuid uint32
	var err error
	var opt int
	inputReader = bufio.NewReader(os.Stdin)
	for {
		fmt.Println("-- 1. Send message --")
		fmt.Println("-- 2. Add friend  --")
		fmt.Println("-- 3. Unaccepted friend invitation  --")
		fmt.Println("-- 4. Chat with friend  --")
		fmt.Println("-- 5. Delete Friend  --")
		fmt.Println("-- 6. Create Group  --")
		fmt.Println("-- 7. Join Group  --")
		fmt.Println("-- 8. View Group  --")
		fmt.Println("-- 9. Send File  --")

		_, _ = fmt.Scanf("%d", &opt)
		//fmt.Print("\n\n\n\n\n\n\n")
		switch opt {
		case 1:
			fmt.Println("input uuid:")
			_, err = fmt.Scanf("%d", &uuid)
			fmt.Println("input message")
			for {
				input, err = inputReader.ReadString('\n')
				if err != nil || len(input) == 0{
					continue
				}
				break
			}
			input = input[:len(input)-1]
			autoAdaptSendMessage(def.Message, uuid, input, cli)
		case 2: // 添加好友
			fmt.Println("Enter '0' to exit")
			fmt.Println("Please enter uuid:")
			_, _ = fmt.Scanf("%d", &uuid)
			if uuid == 0 {
				continue
			}
			data := NewPackage()
			data.SetRequestCode(def.AddFriend)
			data.SetHeadData(UInt32ToBytes(uuid))
			cli.DataSend <- &data
		case 3: // 查看未接受的好友邀请
			fmt.Println("Enter '0' to exit")
			fmt.Println("Please wait......")
			data := NewPackage()
			data.SetRequestCode(def.ListFriendInvitation)
			cli.DataSend <- &data

			_, err = fmt.Scanf("%d", &uuid)
			if err != nil || uuid == 0 {
				continue
			}
			data = NewPackage()
			data.SetRequestCode(def.AcceptFriendInvitation)
			data.SetHeadData(UInt32ToBytes(uuid))
			cli.DataSend <- &data
		case 4: // 和好友聊天
			fmt.Println("Enter '0' to exit")
			fmt.Println("Please wait......")
			data := NewPackage()
			data.SetRequestCode(def.ListFriend)
			cli.DataSend <- &data

			_, err = fmt.Scanf("%d", &uuid)
			if err != nil || uuid == 0 {
				continue
			}

			fmt.Println("Enter '.exit' to exit")
			fmt.Println("Enter your message")

			for {
				input, err = inputReader.ReadString('\n')
				if err != nil || len(input) == 0{
					continue
				}
				input = input[:len(input)-1]
				if input == ".exit" {
					break
				}
				autoAdaptSendMessage(def.Message, uuid, input, cli)
			}
		case 5: // 删除好友
			fmt.Println("Enter '0' to exit")
			fmt.Println("Please wait......")
			data := NewPackage()
			data.SetRequestCode(def.ListFriend)
			cli.DataSend <- &data

			_, err = fmt.Scanf("%d", &uuid)
			if err != nil || uuid == 0 {
				continue
			}
			data = NewPackage()
			data.SetRequestCode(def.DeleteFriend)
			data.SetHeadData(UInt32ToBytes(uuid))
			cli.DataSend <- &data
		case 6: // 创建群组
			fmt.Println("Enter '.exit' to exit")
			fmt.Println("Please Enter Group name")
			input, err = inputReader.ReadString('\n')
			if err != nil || len(input) == 0{
				continue
			}
			input = input[:len(input)-1]
			if input == ".exit" {
				break
			}
			autoAdaptSendMessage(def.CreateGroup, cli.UUID, input, cli)
		case 7: // 加入群组
			fmt.Println("Enter '0' to exit")
			fmt.Println("Please wait......")
			_, err = fmt.Scanf("%d", &uuid)
			if err != nil || uuid == 0 {
				continue
			}
			data := NewPackage()
			data.SetRequestCode(def.JoinGroup)
			data.SetHeadData(UInt32ToBytes(uuid))
			cli.DataSend <- &data
		case 8:
			fmt.Println("Enter '0' to exit")
			fmt.Println("Please wait......")
			data := NewPackage()
			data.SetRequestCode(def.ListGroup)
			cli.DataSend <- &data
			guid := uint32(0)
			_, err = fmt.Scanf("%d", &guid)
			if err != nil || guid == 0 {
				continue
			}
			GROUP:
			fmt.Println("*********************************************")
			fmt.Println("* 1. Chat in group                          *")
			fmt.Println("* 2. Send File to group                     *")
			fmt.Println("* 3. Manage group application (owner/admin) *")
			fmt.Println("* 4. Appoint administrator (owner only)     *")
			fmt.Println("* 5. Revoke administrator (owner only)      *")
			fmt.Println("* 6. Transfer ownership (owner only)        *")
			fmt.Println("* 7. Delete member (owner/admin)            *")
			fmt.Println("* 8. Download Group File                    *")
			fmt.Println("*                                           *")
			fmt.Println("* 0. Exit                                   *")
			fmt.Println("*********************************************")
			opt := uint32(0)
			_, err = fmt.Scanf("%d", &opt)
			if err != nil || opt == 0 {
				continue
			}
			switch opt {
			case 1: // Chat in group
				for {
					input, err = inputReader.ReadString('\n')
					if err != nil || len(input) == 0{
						continue
					}
					input = input[:len(input)-1]
					if input == ".exit" {
						break
					}
					autoAdaptSendGroupMessage(guid, input, cli)
				}
				goto GROUP
			case 2: // Send File to group
				//// TODO Send File to group
				//fmt.Println("Please enter guid:")
				//_, err = fmt.Scanf("%d", &guid)
				//if err != nil || uuid == 0 {
				//	continue
				//}
				fmt.Println("Please enter the absolute file path")
				input, err = inputReader.ReadString('\n')
				if err != nil || len(input) == 0{
					goto GROUP
				}
				input = input[:len(input)-1]
				if input == ".exit" {
					goto GROUP
				}
				fmt.Println("Calculating file hash value")
				filename, filePath, hashVal := CalculateFileHashValue(input)
				fmt.Printf("Filename: [%s] Path:[%s] Hash:[%d]\n", filename, filePath, hashVal)
				fmt.Println("Preparing file")
				FileSendPrepare(filename, filePath, cli.UUID, guid, hashVal)

				data := NewPackage()
				mes, ok := WrapMessage(guid, filename)
				if !ok {
					fmt.Println("Filename to long")
					goto GROUP
				}
				data.SetHeadData(mes)
				data.SetExternalDataCheckSum(hashVal)
				data.SetExtendedDataFlag(1)
				data.SetRequestCode(def.SendGroupFile)
				cli.DataSend <- &data
				goto GROUP
			case 3: // Manage group application
				fmt.Println("Please wait......")
				data := NewPackage()
				data.SetHeadData(UInt32ToBytes(guid))
				data.SetRequestCode(def.ListJoinGroup)
				cli.DataSend <- &data
				_, err = fmt.Scanf("%d", &uuid)
				if err != nil || uuid == 0 {
					goto GROUP
				}
				data = NewPackage()
				data.SetHeadData(WrapGuidUuid(guid, uuid))
				data.SetRequestCode(def.AcceptJoinGroup)
				cli.DataSend <- &data
				goto GROUP
			case 4: // Appoint Administrator
				fmt.Println("Please wait......")
				data := NewPackage()
				data.SetHeadData(UInt32ToBytes(guid))
				data.SetRequestCode(def.ListGroupMember)
				cli.DataSend <- &data
				_, err = fmt.Scanf("%d", &uuid)
				if err != nil || uuid == 0 {
					goto GROUP
				}
				data = NewPackage()
				data.SetHeadData(WrapGuidUuid(guid, uuid))
				data.SetRequestCode(def.AppointAdmin)
				cli.DataSend <- &data
				goto GROUP
			case 5: // Revoke Administrator
				fmt.Println("Please wait......")
				data := NewPackage()
				data.SetHeadData(UInt32ToBytes(guid))
				data.SetRequestCode(def.ListGroupAdmin)
				cli.DataSend <- &data
				_, err = fmt.Scanf("%d", &uuid)
				if err != nil || uuid == 0 {
					goto GROUP
				}
				data = NewPackage()
				data.SetHeadData(WrapGuidUuid(guid, uuid))
				data.SetRequestCode(def.RevokeAdmin)
				cli.DataSend <- &data
				goto GROUP
			case 6: // Transfer ownership
				fmt.Println("Please wait......")
				data := NewPackage()
				data.SetHeadData(UInt32ToBytes(guid))
				data.SetRequestCode(def.ListGroupMember)
				cli.DataSend <- &data
				_, err = fmt.Scanf("%d", &uuid)
				if err != nil || uuid == 0 {
					goto GROUP
				}
				data = NewPackage()
				data.SetHeadData(WrapGuidUuid(guid, uuid))
				data.SetRequestCode(def.TransferGroup)
				cli.DataSend <- &data
				goto GROUP
			case 7: // Delete member
				fmt.Println("Please wait......")
				data := NewPackage()
				data.SetHeadData(UInt32ToBytes(guid))
				data.SetRequestCode(def.ListGroupMember)
				cli.DataSend <- &data
				_, err = fmt.Scanf("%d", &uuid)
				if err != nil || uuid == 0 {
					goto GROUP
				}
				data = NewPackage()
				data.SetHeadData(WrapGuidUuid(guid, uuid))
				data.SetRequestCode(def.DeleteMember)
				cli.DataSend <- &data
				goto GROUP
			case 8: // Download Group File
				fmt.Println("Please wait......")
				data := NewPackage()
				data.SetHeadData(UInt32ToBytes(guid))
				data.SetRequestCode(def.ListGroupFile)
				cli.DataSend <- &data
				fuid := uint32(0)
				_, err = fmt.Scanf("%d", &fuid)
				if err != nil || fuid == 0 {
					goto GROUP
				}
				data = NewPackage()
				data.SetHeadData(WrapGuidUuid(guid, fuid))
				data.SetRequestCode(def.DownloadGroupFile)
				cli.DataSend <- &data
			default:
				goto GROUP
			}
		case 9:
			fmt.Println("Please enter uuid:")
			_, err = fmt.Scanf("%d", &uuid)
			if err != nil || uuid == 0 {
				continue
			}
			fmt.Println("Please enter the absolute file path")
			input, err = inputReader.ReadString('\n')
			if err != nil || len(input) == 0{
				continue
			}
			input = input[:len(input)-1]
			if input == ".exit" {
				break
			}
			fmt.Println("Calculating file hash value")
			filename, filePath, hashVal := CalculateFileHashValue(input)
			fmt.Printf("Filename: [%s] Path:[%s] Hash:[%d]\n", filename, filePath, hashVal)
			fmt.Println("Preparing file")
			FileSendPrepare(filename, filePath, cli.UUID, uuid, hashVal)

			data := NewPackage()
			mes, ok := WrapMessage(uuid, filename)
			if !ok {
				fmt.Println("Filename to long")
				continue
			}
			data.SetHeadData(mes)
			data.SetExternalDataCheckSum(hashVal)
			data.SetExtendedDataFlag(1)
			data.SetRequestCode(def.SendFile)
			cli.DataSend <- &data
		}
	}
}

func autoAdaptSendMessage(requestCode uint16, uuid uint32, txt string, cli *Connection) {
	data := NewPackage()
	data.SetRequestCode(requestCode)

	message, ok := WrapMessage(uuid, txt)
	if !ok {
		message, _ = WrapLongMessage(uuid, txt)
		data.SetRequestCode(requestCode)
		data.SetExtendedDataFlag(1)
		data.SetExternalDataCheckSum(CRC32([]byte(txt)))
	}

	data.SetHeadData(message)
	cli.DataSend <- &data
}

func autoAdaptSendGroupMessage(guid uint32, txt string, cli *Connection) {
	data := NewPackage()
	data.SetRequestCode(def.GroupMessage)

	message, ok := WrapGroupMessage(guid, cli.UUID, txt)
	if !ok {
		message, _ = WrapLongGroupMessage(guid, cli.UUID, txt)
		data.SetRequestCode(def.GroupMessage)
		data.SetExtendedDataFlag(1)
		data.SetExternalDataCheckSum(CRC32([]byte(txt)))
	}
	data.SetHeadData(message)
	cli.DataSend <- &data
}