package sockssl

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

const (
	AddrTypeIPv4   = 1
	AddrTypeIPv6   = 4
	AddrTypeDomain = 3
)

func DecodeAddress(c net.Conn, saveRaw bool) (addr string, raw []byte, err error) {
	var host, port string
	var buf []byte
	buf, err = readBytes(c, 1)
	if err != nil {
		return
	}
	if saveRaw {
		raw = append(raw, buf...)
	}
	var addrLen int
	atype := buf[0]
	switch atype {
	case AddrTypeIPv4:
		addrLen = net.IPv4len
	case AddrTypeIPv6:
		addrLen = net.IPv6len
	case AddrTypeDomain:
		buf, err = readBytes(c, 1)
		if err != nil {
			return
		}
		if saveRaw {
			raw = append(raw, buf...)
		}
		addrLen = int(buf[0])
	default:
		err = fmt.Errorf("unknown address type %x", atype)
		return
	}
	buf, err = readBytes(c, addrLen+2) // 2 bytes for port
	if err != nil {
		return
	}
	if saveRaw {
		raw = append(raw, buf...)
	}
	if atype == AddrTypeIPv4 || atype == AddrTypeIPv6 {
		host = net.IP(buf[:addrLen]).String()
	} else {
		host = string(buf[:addrLen])
	}
	port = strconv.Itoa(int(buf[addrLen])<<8 + int(buf[addrLen+1]))
	addr = net.JoinHostPort(host, port)
	return
}

func Handshake(c net.Conn) (addr string, raw []byte, err error) {
	if err = handleAuthMethods(c); err != nil {
		return
	}
	if err = handleConnectCommand(c); err != nil {
		return
	}
	if addr, raw, err = DecodeAddress(c, true); err != nil {
		return
	}
	if _, err = c.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}); err != nil {
		return
	}
	return
}

func handleAuthMethods(c net.Conn) (err error) {
	buf := make([]byte, 2)
	if _, err = io.ReadFull(c, buf[:2]); err != nil {
		return
	}
	buf = make([]byte, buf[1])
	if _, err = io.ReadFull(c, buf); err != nil {
		return
	}
	if !bytes.Contains(buf, []byte{0}) {
		return errors.New("no authentication required")
	}
	if _, err = c.Write([]byte{5, 0}); err != nil {
		return
	}
	return
}

func handleConnectCommand(c net.Conn) (err error) {
	buf := make([]byte, 3)
	_, err = io.ReadFull(c, buf[:3])
	if err != nil {
		return
	}
	if buf[1] != 1 {
		return fmt.Errorf("unsupported CMD %x", buf)
	}
	return
}

func readBytes(c net.Conn, n int) (buf []byte, err error) {
	buf = make([]byte, n)
	_, err = io.ReadFull(c, buf)
	return
}

func IOCopyLoop(c1, c2 net.Conn) (sent, recv int64) {
	ch := make(chan int64)
	go func() {
		n, _ := io.Copy(c1, c2)
		c1.Close()
		ch <- n
	}()
	sent, _ = io.Copy(c2, c1)
	c2.Close()
	recv = <-ch
	return
}
