package uuid

import (
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"
)

const encodeStr = "123456789abcdefghijklmnpqrstuvxy"

var (
	seq      uint32
	initBuf  = make([]byte, 14)
	encoding = base32.NewEncoding(encodeStr).WithPadding(base32.NoPadding)
)

func init() {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if i, ok := addr.(*net.IPNet); ok {
			if !i.IP.IsLoopback() {
				i4 := i.IP.To4()
				if len(i4) == net.IPv4len {
					initBuf[4] = i.IP.To4()[0]
					initBuf[5] = i.IP.To4()[1]
					initBuf[6] = i.IP.To4()[2]
					initBuf[7] = i.IP.To4()[3]
					break
				}
			}
		}
	}
	binary.LittleEndian.PutUint16(initBuf[8:], uint16(os.Getpid()))
}

//String 生成字符串的uuid
func String() string {
	buf := initBuf[:]
	binary.BigEndian.PutUint32(buf[0:], uint32(time.Now().UTC().Unix()))
	binary.BigEndian.PutUint32(buf[10:], atomic.AddUint32(&seq, 1))
	return encoding.EncodeToString(buf)
}

//Decode 解析uuid字符串，返回具体细节.
func Decode(s string) (ip string, pid int, tm time.Time, seq uint32, err error) {
	var buf []byte
	buf, err = encoding.DecodeString(s)
	if err != nil {
		return
	}

	tm = time.Unix(int64(binary.BigEndian.Uint32(buf)), 0)
	ip = net.IPv4(buf[4], buf[5], buf[6], buf[7]).String()
	pid = int(binary.LittleEndian.Uint16(buf[8:]))
	seq = binary.BigEndian.Uint32(buf[10:])

	return
}

//Info 解析uuid中信息.
func Info(s string) (string, error) {
	ip, pid, tm, seq, err := Decode(s)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ip:%s pid:%d time:%s sequence:%d", ip, pid, tm.Format(time.RFC3339), seq), nil
}
