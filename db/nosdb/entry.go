package nosdb

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"time"
)

// 4 + 8 + 4 + 4 + 4 + 2 + 2 + 2
const ENTRY_MERA_SIZE = 30

type MARK uint16

const (
	PUT MARK = iota
	DEL

	Persistent uint32 = 0
)

// entry 的一些`meta`信息
type Meta struct {
	CRC       uint32   // crc 校验信息
	Timestamp uint64   // 时间戳
	TTL       uint32   // 定时时间, 0 代表用不过期
	KeySize   uint32   // key 的长度
	ValueSize uint32   // value 长度
	Mark      MARK     // 标记，PUT or DEL
	Encoding  ENCODING // 底层数据结构，编码类型
	Ty        TYPE     // 存储类型， list, set, zset
}

// 一个存储条目
//
//  the entry stored format:
//  |----------------------------------------------------------------------------------------------------------------|
//  |  crc  | timestamp | keysize | valueSize | mark   | TTL  |bucketSize| status | ds   | txId |  bucket |  key  | value |
//  |----------------------------------------------------------------------------------------------------------------|
//  | uint32| uint64    |uint32   |  uint32   | uint16 | uint32| uint32 | uint16 | uint16 |uint64 |[]byte|[]byte | []byte |
//  |----------------------------------------------------------------------------------------------------------------|
type Entry struct {
	Meta
	Key   []byte
	Value []byte
}

// 新建一条记录
func NewEntry(key, value []byte, mark MARK, TTL uint32, encoding ENCODING, ty TYPE) *Entry {
	return &Entry{
		Key:   key,
		Value: value,
		Meta: Meta{
			Timestamp: uint64(time.Now().Unix()),
			TTL:       TTL,
			KeySize:   uint32(len(key)),
			ValueSize: uint32(len(value)),
			Mark:      mark,
			Encoding:  encoding,
			Ty:        ty,
		},
	}
}

// buf 要检验的字节流
// 返回 校验结果
func (e *Entry) GetCRC(buf []byte) uint32 {
	return crc32.ChecksumIEEE(buf)
}

func (e *Entry) GetSize() int64 {
	return (int64)(ENTRY_MERA_SIZE + e.KeySize + e.ValueSize)
}

// encode 编码，返回字节数组
func (e *Entry) Encode() ([]byte, error) {
	buf := make([]byte, e.GetSize())
	binary.BigEndian.PutUint64(buf[4:12], e.Timestamp)
	binary.BigEndian.PutUint32(buf[12:16], e.TTL)
	binary.BigEndian.PutUint32(buf[16:20], e.KeySize)
	binary.BigEndian.PutUint32(buf[20:24], e.ValueSize)
	binary.BigEndian.PutUint32(buf[20:24], e.ValueSize)
	binary.BigEndian.PutUint16(buf[24:26], uint16(e.Mark))
	binary.BigEndian.PutUint16(buf[26:28], uint16(e.Ty))
	binary.BigEndian.PutUint16(buf[28:30], uint16(e.Encoding))
	copy(buf[ENTRY_MERA_SIZE:ENTRY_MERA_SIZE+e.KeySize], e.Key)
	copy(buf[ENTRY_MERA_SIZE+e.KeySize:], e.Value)
	binary.BigEndian.PutUint32(buf[0:4], e.GetCRC(buf[4:]))
	return buf, nil
}

// 将字节流解码为 entry 实体
func DecodeMeta(buf []byte) (entry *Entry, err error) {
	if len(buf) != ENTRY_MERA_SIZE {
		err = fmt.Errorf(" len is not match ")
		return
	}
	entry = &Entry{}
	entry.CRC = binary.BigEndian.Uint32(buf[0:4])
	entry.Timestamp = binary.BigEndian.Uint64(buf[4:12])
	entry.TTL = binary.BigEndian.Uint32(buf[12:16])
	entry.KeySize = binary.BigEndian.Uint32(buf[16:20])
	entry.ValueSize = binary.BigEndian.Uint32(buf[20:24])
	entry.Mark = (MARK)(binary.BigEndian.Uint32(buf[24:26]))
	entry.Ty = (TYPE)(binary.BigEndian.Uint32(buf[26:28]))
	entry.Encoding = (ENCODING)(binary.BigEndian.Uint32(buf[28:30]))
	return
}

func (e *Entry) CheckCRC(buf []byte) bool {
	// 更新新的crc
	crc := e.GetCRC(buf)
	return crc == e.CRC
}