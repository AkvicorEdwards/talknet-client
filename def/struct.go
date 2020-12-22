package def

type Files struct {
	Fuid uint32 `json:"fuid"` // 文件id
	Uuid uint32 `json:"uuid"` // 上传者
	Filename string `json:"filename"`
	RealName string `json:"real_name"`
	Hash uint32 `json:"hash"`
}
