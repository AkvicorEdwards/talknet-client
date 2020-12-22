package tcp

const (
	// 数据最大长度
	LengthHeadPackage     = 210
	// 偏移
	OffsetRequestCode = 0
	LengthRequestCode = 2
	OffsetSEQ = 2
	LengthSEQ = 4
	OffsetACK = 6
	LengthACK = 4
	OffsetTime = 10
	LengthTime = 8
	OffsetExtendedDataFlag = 18
	LengthExtendedDataFlag = 1
	OffsetHeadDataLength = 19
	LengthHeadDataLength = 1
	OffsetHeadData = 20
	LengthHeadData = 182
	OffsetExtendedDataHash = 202
	LengthExtendedDataHash = 4
	OffsetHash = 206
	LengthHash = 4
)


type Package struct {
	data [LengthHeadPackage]byte
}

// 将byte数组转化为Package
func ConvertToPackage(data []byte) (p Package) {
	if len(data) < LengthHeadPackage {
		copy(p.data[:], data[:])
	}
	copy(p.data[:], data[:LengthHeadPackage])
	return
}

func NewPackage() Package {
	return Package{data: [LengthHeadPackage]byte{}}
}

// 获取data的内容
func (p *Package) Data() []byte {
	return p.data[:]
}

func (p *Package) SetRequestCode(request uint16) {
	data := UInt16ToBytes(request)
	for k, v := range data {
		p.data[OffsetRequestCode+k] = v
	}
}

func (p *Package) GetRequestCode() uint16 {
	return BytesToUInt16(p.data[:OffsetRequestCode+LengthRequestCode])
}

func (p *Package) ClearRequestCode() {
	for i := OffsetRequestCode; i < OffsetRequestCode+LengthRequestCode; i++ {
		p.data[i] = 0
	}
}

func (p *Package) SetSEQ(seq uint32) {
	data := UInt32ToBytes(seq)
	for k, v := range data {
		p.data[OffsetSEQ+k] = v
	}
}

func (p *Package) GetSEQ() uint32 {
	return BytesToUInt32(p.data[OffsetSEQ:OffsetSEQ+LengthSEQ])
}

func (p *Package) ClearSEQ() {
	for i := OffsetSEQ; i < OffsetSEQ+LengthSEQ; i++ {
		p.data[i] = 0
	}
}

func (p *Package) SetACK(ack uint32) {
	data := UInt32ToBytes(ack)
	for k, v := range data {
		p.data[OffsetACK+k] = v
	}
}

func (p *Package) GetACK() uint32 {
	return BytesToUInt32(p.data[OffsetACK:OffsetACK+LengthACK])
}

func (p *Package) ClearACK() {
	for i := OffsetACK; i < OffsetACK+LengthACK; i++ {
		p.data[i] = 0
	}
}

func (p *Package) SetTime(t uint64) {
	data := UInt64ToBytes(t)
	for k, v := range data {
		p.data[OffsetTime+k] = v
	}
}

func (p *Package) GetTime() uint64 {
	return BytesToUInt64(p.data[OffsetTime:OffsetTime+LengthTime])
}

func (p *Package) ClearTime() {
	for i := OffsetTime; i < OffsetTime+LengthTime; i++ {
		p.data[i] = 0
	}
}

func (p *Package) SetExtendedDataFlag(flag byte) {
	p.data[OffsetExtendedDataFlag] = flag
}

func (p *Package) GetExtendedDataFlag() byte {
	return p.data[OffsetExtendedDataFlag]
}

func (p *Package) ClearExtendedDataFlag() {
	for i := OffsetExtendedDataFlag; i < OffsetExtendedDataFlag+LengthExtendedDataFlag; i++ {
		p.data[i] = 0
	}
}

func (p *Package) SetHeadData(data []byte) {
	if len(data) > LengthHeadData {
		data = data[:LengthHeadData]
	}
	p.data[OffsetHeadDataLength] = byte(len(data))
	for k, v := range data {
		p.data[OffsetHeadData+k] = v
	}
}

func (p *Package) GetHeadData() []byte {
	return p.data[OffsetHeadData:OffsetHeadData+p.data[OffsetHeadDataLength]]
}

func (p *Package) ClearHeadData() {
	for i := OffsetHeadDataLength; i < OffsetHeadDataLength+LengthHeadDataLength; i++ {
		p.data[i] = 0
	}
	for i := OffsetHeadData; i < OffsetHeadData+LengthHeadData; i++ {
		p.data[i] = 0
	}
}

func (p *Package) SetExternalDataCheckSum(sum uint32) {
	data := UInt32ToBytes(sum)
	for k, v := range data {
		p.data[OffsetExtendedDataHash+k] = v
	}
}

func (p *Package) GetExternalDataCheckSum() uint32 {
	return BytesToUInt32(p.data[OffsetExtendedDataHash:OffsetExtendedDataHash+LengthExtendedDataHash])
}

func (p *Package) ClearExternalDataCheckSum() {
	for i := OffsetExtendedDataHash; i < OffsetExtendedDataHash+LengthExtendedDataHash; i++ {
		p.data[i] = 0
	}
}

func (p *Package) CheckExternalDataCheckSum(sum uint32) bool {
	return sum == p.GetExternalDataCheckSum()
}


func (p *Package) SetHeadCheckSum() {
	data := UInt32ToBytes(CRC32(p.data[:OffsetHash]))
	for k, v := range data {
		p.data[OffsetHash+k] = v
	}
}

func (p *Package) GetHeadCheckSum() uint32 {
	return BytesToUInt32(p.data[OffsetHash:OffsetHash+LengthHash])
}

func (p *Package) ClearHeadCheckSum() {
	for i := OffsetHash; i < OffsetHash+LengthHash; i++ {
		p.data[i] = 0
	}
}

func (p *Package) CheckHeadCheckSum() bool {
	return CRC32(p.data[:OffsetHash]) == p.GetHeadCheckSum()
}


func (p *Package) Clear() {
	p.ClearRequestCode()
	p.ClearSEQ()
	p.ClearACK()
	p.ClearTime()
	p.ClearExtendedDataFlag()
	p.ClearHeadData()
	p.ClearHeadCheckSum()
	p.ClearExternalDataCheckSum()
}

func (p *Package) ClearExceptSeq() {
	p.ClearRequestCode()
	p.ClearACK()
	p.ClearTime()
	p.ClearExtendedDataFlag()
	p.ClearHeadData()
	p.ClearHeadCheckSum()
	p.ClearExternalDataCheckSum()
}
