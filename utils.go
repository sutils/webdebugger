package webdebugger

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/websocket"
)

//BasePipe is func to create os pipe
var BasePipe = os.Pipe

//PipedConn is connection piped read and write
type PipedConn struct {
	io.Reader
	io.Writer
	Alias string
}

//CreatePipeConn will create pipe connection
func CreatePipeConn() (a, b *PipedConn, err error) {
	aReader, bWriter, err := BasePipe()
	if err != nil {
		return
	}
	bReader, aWriter, err := BasePipe()
	if err != nil {
		aReader.Close()
		bWriter.Close()
		return
	}
	a = &PipedConn{
		Reader: aReader,
		Writer: aWriter,
		Alias:  fmt.Sprintf("%v,%v", aReader, aWriter),
	}
	b = &PipedConn{
		Reader: bReader,
		Writer: bWriter,
		Alias:  fmt.Sprintf("%v,%v", bReader, bWriter),
	}
	return
}

//Close will close Reaer/Writer
func (p *PipedConn) Close() (err error) {
	if closer, ok := p.Reader.(io.Closer); ok {
		err = closer.Close()
	}
	if closer, ok := p.Writer.(io.Closer); ok {
		xerr := closer.Close()
		if err == nil {
			err = xerr
		}
	}
	return
}

//Network is net.Addr impl
func (p *PipedConn) Network() string {
	return "Piped"
}

//LocalAddr is net.Conn impl
func (p *PipedConn) LocalAddr() net.Addr {
	return p
}

//RemoteAddr is net.Conn impl
func (p *PipedConn) RemoteAddr() net.Addr {
	return p
}

//SetDeadline is net.Conn impl
func (p *PipedConn) SetDeadline(t time.Time) error {
	return nil
}

//SetReadDeadline is net.Conn impl
func (p *PipedConn) SetReadDeadline(t time.Time) error {
	return nil
}

//SetWriteDeadline is net.Conn impl
func (p *PipedConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (p *PipedConn) String() string {
	return p.Alias
}

const (
	//LogLevelDebug is debug log level
	LogLevelDebug = 40
	//LogLevelInfo is info log level
	LogLevelInfo = 30
	//LogLevelWarn is warn log level
	LogLevelWarn = 20
	//LogLevelError is error log level
	LogLevelError = 10
)

var logLevel = LogLevelInfo

//SetLogLevel is set log level to l
func SetLogLevel(l int) {
	if l > 0 {
		logLevel = l
	}
}

//DebugLog is the debug level log
func DebugLog(format string, args ...interface{}) {
	if logLevel < LogLevelDebug {
		return
	}
	log.Output(2, fmt.Sprintf("D "+format, args...))
}

//InfoLog is the info level log
func InfoLog(format string, args ...interface{}) {
	if logLevel < LogLevelInfo {
		return
	}
	log.Output(2, fmt.Sprintf("I "+format, args...))
}

//WarnLog is the warn level log
func WarnLog(format string, args ...interface{}) {
	if logLevel < LogLevelWarn {
		return
	}
	log.Output(2, fmt.Sprintf("W "+format, args...))
}

//ErrorLog is the error level log
func ErrorLog(format string, args ...interface{}) {
	if logLevel < LogLevelError {
		return
	}
	log.Output(2, fmt.Sprintf("E "+format, args...))
}

//ReadJSON will read file and unmarshal to value
func ReadJSON(filename string, v interface{}) (err error) {
	data, err := ioutil.ReadFile(filename)
	if err == nil {
		err = json.Unmarshal(data, v)
	}
	return
}

//SHA1 will get sha1 hash of data
func SHA1(data []byte) string {
	s := sha1.New()
	s.Write(data)
	return fmt.Sprintf("%x", s.Sum(nil))
}

//StringConn is an ReadWriteCloser for return  remote address info
type StringConn struct {
	Name string
	net.Conn
}

//NewStringConn will return new StringConn
func NewStringConn(raw net.Conn) *StringConn {
	return &StringConn{
		Conn: raw,
	}
}

func (s *StringConn) String() string {
	if len(s.Name) > 0 {
		return s.Name
	}
	return remoteAddr(s.Conn)
}

func remoteAddr(v interface{}) string {
	if wsc, ok := v.(*websocket.Conn); ok {
		return fmt.Sprintf("%v", wsc.RemoteAddr())
	}
	if netc, ok := v.(net.Conn); ok {
		return fmt.Sprintf("%v", netc.RemoteAddr())
	}
	return fmt.Sprintf("%v", v)
}

//TCPKeepAliveListener is normal tcp listner for set tcp connection keep alive
type TCPKeepAliveListener struct {
	*net.TCPListener
}

//Accept will accept one connection
func (ln TCPKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err == nil {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(3 * time.Minute)
	}
	return tc, err
}

func httpGet(url, username, password string) (data []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err == nil {
		if len(username) > 0 {
			req.SetBasicAuth(username, password)
		}
		var resp *http.Response
		resp, err = http.DefaultClient.Do(req)
		if err == nil {
			data, err = ioutil.ReadAll(resp.Body)
		}
	}
	return
}
