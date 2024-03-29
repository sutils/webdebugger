package webdebugger

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func proxyDial(t *testing.T, remote string, port uint16) {
	conn, err := net.Dial("tcp", "localhost:2081")
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()
	buf := make([]byte, 1024*64)
	proxyReader := bufio.NewReader(conn)
	_, err = conn.Write([]byte{0x05, 0x01, 0x00})
	if err != nil {
		return
	}
	err = fullBuf(proxyReader, buf, 2)
	if err != nil {
		return
	}
	if buf[0] != 0x05 || buf[1] != 0x00 {
		err = fmt.Errorf("only ver 0x05 / method 0x00 is supported, but %x/%x", buf[0], buf[1])
		return
	}
	buf[0], buf[1], buf[2], buf[3] = 0x05, 0x01, 0x00, 0x03
	buf[4] = byte(len(remote))
	copy(buf[5:], []byte(remote))
	binary.BigEndian.PutUint16(buf[5+len(remote):], port)
	_, err = conn.Write(buf[:buf[4]+7])
	if err != nil {
		return
	}
	readed, err := proxyReader.Read(buf)
	if err != nil {
		return
	}
	fmt.Printf("->%v\n", buf[0:readed])
}

func proxyDial2(t *testing.T, remote string, port uint16) {
	conn, err := net.Dial("tcp", "localhost:2081")
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()
	buf := make([]byte, 1024*64)
	proxyReader := bufio.NewReader(conn)
	_, err = conn.Write([]byte{0x05, 0x01, 0x00})
	if err != nil {
		return
	}
	err = fullBuf(proxyReader, buf, 2)
	if err != nil {
		return
	}
	if buf[0] != 0x05 || buf[1] != 0x00 {
		err = fmt.Errorf("only ver 0x05 / method 0x00 is supported, but %x/%x", buf[0], buf[1])
		return
	}
	buf[0], buf[1], buf[2], buf[3] = 0x05, 0x01, 0x00, 0x13
	buf[4] = byte(len(remote))
	copy(buf[5:], []byte(remote))
	binary.BigEndian.PutUint16(buf[5+len(remote):], port)
	_, err = conn.Write(buf[:buf[4]+7])
	if err != nil {
		return
	}
	readed, err := proxyReader.Read(buf)
	if err != nil {
		return
	}
	fmt.Printf("->%v\n", buf[0:readed])
}

func proxyDialIP(t *testing.T, bys []byte, port uint16) {
	conn, err := net.Dial("tcp", "localhost:2081")
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()
	buf := make([]byte, 1024*64)
	proxyReader := bufio.NewReader(conn)
	_, err = conn.Write([]byte{0x05, 0x01, 0x00})
	if err != nil {
		return
	}
	err = fullBuf(proxyReader, buf, 2)
	if err != nil {
		return
	}
	if buf[0] != 0x05 || buf[1] != 0x00 {
		err = fmt.Errorf("only ver 0x05 / method 0x00 is supported, but %x/%x", buf[0], buf[1])
		return
	}
	buf[0], buf[1], buf[2], buf[3] = 0x05, 0x01, 0x00, 0x01
	copy(buf[4:], bys)
	binary.BigEndian.PutUint16(buf[8:], port)
	_, err = conn.Write(buf[:10])
	if err != nil {
		return
	}
	readed, err := proxyReader.Read(buf)
	if err != nil {
		return
	}
	fmt.Printf("->%v\n", buf[0:readed])
}

func proxyDialIPv6(t *testing.T, bys []byte, port uint16) {
	conn, err := net.Dial("tcp", "localhost:2081")
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()
	buf := make([]byte, 1024*64)
	proxyReader := bufio.NewReader(conn)
	_, err = conn.Write([]byte{0x05, 0x01, 0x00})
	if err != nil {
		return
	}
	err = fullBuf(proxyReader, buf, 2)
	if err != nil {
		return
	}
	if buf[0] != 0x05 || buf[1] != 0x00 {
		err = fmt.Errorf("only ver 0x05 / method 0x00 is supported, but %x/%x", buf[0], buf[1])
		return
	}
	buf[0], buf[1], buf[2], buf[3] = 0x05, 0x01, 0x00, 0x04
	copy(buf[4:], bys)
	binary.BigEndian.PutUint16(buf[8:], port)
	_, err = conn.Write(buf[:10])
	if err != nil {
		return
	}
	readed, err := proxyReader.Read(buf)
	if err != nil {
		return
	}
	fmt.Printf("->%v\n", buf[0:readed])
}

type CodableErr struct {
	Err error
}

func (c *CodableErr) Error() string {
	return c.Err.Error()
}

func (c *CodableErr) Code() byte {
	return 0x01
}

func TestSocksProxy(t *testing.T) {
	proxy := NewSocksProxy()
	proxy.ProcConn = func(uri string, raw net.Conn) (async bool, err error) {
		conn, err := net.Dial("tcp", uri)
		if err == nil {
			go io.Copy(conn, raw)
			go io.Copy(raw, conn)
			time.Sleep(100 * time.Millisecond)
		} else {
			err = &CodableErr{Err: err}
		}
		fmt.Println("dial to ", uri, err)
		return
	}
	go func() {
		err := proxy.Listen(":2081")
		if err != nil {
			t.Error(err)
			return
		}
		proxy.Run()
	}()
	proxyDial(t, "localhost", 80)
	proxyDial2(t, "localhost:80", 0)
	proxyDial(t, "localhost", 81)
	proxyDialIP(t, make([]byte, 4), 80)
	proxyDialIPv6(t, make([]byte, 16), 80)
	{ //test error
		//
		conn, conb, _ := CreatePipeConn()
		go proxy.procConn(conb)
		conn.Close()
		//
		conn, conb, _ = CreatePipeConn()
		go proxy.procConn(conb)
		conn.Write([]byte{0x00, 0x00})
		conn.Close()
		//
		conn, conb, _ = CreatePipeConn()
		go proxy.procConn(conb)
		conn.Write([]byte{0x05, 0x01})
		conn.Close()
		//
		conn, conb, _ = CreatePipeConn()
		go proxy.procConn(conb)
		conn.Write([]byte{0x05, 0x01, 0x00})
		conn.Close()
		//
		conn, conb, _ = CreatePipeConn()
		go proxy.procConn(conb)
		conn.Write([]byte{0x05, 0x01, 0x00})
		conn.Read(make([]byte, 1024))
		conn.Close()
		//
		conn, conb, _ = CreatePipeConn()
		go proxy.procConn(conb)
		conn.Write([]byte{0x05, 0x01, 0x00})
		conn.Read(make([]byte, 1024))
		conn.Write([]byte{0x00, 0x01, 0x00, 0x00, 0x00})
		conn.Close()
		//
		conn, conb, _ = CreatePipeConn()
		go proxy.procConn(conb)
		conn.Write([]byte{0x05, 0x01, 0x00})
		conn.Read(make([]byte, 1024))
		buf := []byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x010}
		binary.BigEndian.PutUint16(buf[8:], 80)
		conn.Write(buf)
		conn.Close()
		time.Sleep(time.Second)
	}
	proxy.Close()
}

func TestFullBuf(t *testing.T) {
	wait := make(chan int, 1)
	r, w, _ := CreatePipeConn()
	go func() {
		buf := make([]byte, 100)
		fullBuf(r, buf, 5)
		fmt.Println("readed", string(buf[:5]))
		wait <- 1
		fullBuf(r, buf, 5)
		wait <- 1
	}()
	w.Write([]byte("abc"))
	time.Sleep(time.Millisecond)
	w.Write([]byte("12"))
	<-wait
	w.Close()
	<-wait
}
