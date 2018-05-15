package uuid

import (
	"encoding/base64"
	"encoding/binary"
	"net"
	"os"
	"sync/atomic"
	"time"
)

var (
	inc uint32
	buf = make([]byte, 16)
)

func init() {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if i, ok := addr.(*net.IPNet); ok {
			if !i.IP.IsLoopback() {
				i4 := i.IP.To4()
				if len(i4) == net.IPv4len {
					buf[0] = i.IP.To4()[2]
					buf[1] = i.IP.To4()[3]
					break
				}
			}
		}
	}
	binary.LittleEndian.PutUint16(buf[2:], uint16(os.Getpid()))
}

//String 生成字符串的uuid
func String() string {
	binary.BigEndian.PutUint64(buf[4:], uint64(time.Now().Unix()))
	binary.BigEndian.PutUint32(buf[12:], atomic.AddUint32(&inc, 1))
	return base64.RawURLEncoding.EncodeToString(buf)
}
