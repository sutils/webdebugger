package webdebugger

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"sync"
)

//Decorder is an interface to decord raw connection by host
type Decorder interface {
	Decord(host string, raw net.Conn) (conn net.Conn, err error)
}

//DecorderCreator is a func define to create Decorder by configure
type DecorderCreator func(name string, config map[string]interface{}) (decorder Decorder, err error)

//DefaultDecorderCreator will create Decorder by name, supported type is TlsDecorder
func DefaultDecorderCreator(name string, config map[string]interface{}) (decorder Decorder, err error) {
	if config == nil {
		err = fmt.Errorf("the %v decorder config is not setted", name)
		return
	}
	decorderType, _ := config["type"].(string)
	switch decorderType {
	case "TlsDecorder":
		d := NewTLSDecorder()
		d.Name, _ = config["name"].(string)
		d.Server, _ = config["server"].(string)
		d.Username, _ = config["username"].(string)
		d.Password, _ = config["password"].(string)
		d.Cert, _ = config["cert"].(string)
		d.Key, _ = config["key"].(string)
		decorder = d
	default:
		err = fmt.Errorf("the %v decorder is not supported", decorderType)
	}
	return
}

//TLSCertCenter provider cert server and it will service the TLSDecorder
type TLSCertCenter struct {
	certs    []map[string]interface{}
	loaded   map[string]*tls.Config
	certsLck sync.RWMutex
}

//NewTLSCertCenter will return new TLSCertCenter by cert configure
func NewTLSCertCenter(certs ...map[string]interface{}) (center *TLSCertCenter) {
	center = &TLSCertCenter{
		certs:    certs,
		loaded:   map[string]*tls.Config{},
		certsLck: sync.RWMutex{},
	}
	return
}

func (t *TLSCertCenter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host, cert, key, ok := t.auth(w, r)
	if !ok {
		return
	}
	_, call := filepath.Split(r.URL.Path)
	switch call {
	// case "decord":
	// 	t.decord(w, r, host, cert, key)
	case "cert":
		t.cert(w, r, host, cert, key)
	default:
		w.WriteHeader(404)
		fmt.Fprintf(w, "%v is not supported", call)
	}
}

func (t *TLSCertCenter) auth(w http.ResponseWriter, r *http.Request) (host, cert, key string, ok bool) {
	host = r.URL.Query().Get("host")
	if len(host) < 1 {
		w.WriteHeader(400)
		fmt.Fprintf(w, "host parameter is requred")
		return
	}
	var conf map[string]interface{}
	var username, password string
	t.certsLck.RLock()
	for _, c := range t.certs {
		h, _ := c["host"].(string)
		if host == h {
			conf = c
			break
		}
	}
	if conf != nil {
		username, _ = conf["username"].(string)
		password, _ = conf["password"].(string)
		cert, _ = conf["cert"].(string)
		key, _ = conf["key"].(string)
	}
	t.certsLck.RUnlock()
	if conf == nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "host config is not exists")
		return
	}
	if len(username) > 0 {
		u, p, _ := r.BasicAuth()
		if u != username || SHA1([]byte(p)) != password {
			w.WriteHeader(401)
			fmt.Fprintf(w, "auth fail")
			return
		}
	}
	ok = true
	return
}

// func (t *TLSCertCenter) decord(w http.ResponseWriter, r *http.Request, host, cert, key string) {
// server := websocket.Server{Handler: func(ws *websocket.Conn) {
// 	config, err := t.tlsConfig(host, cert, key)
// 	if err != nil {
// 		WarnLog("TlsCertCenter load tls config by host:%v,cert:%v,key:%v fail with %v", host, cert, key, err)
// 		return
// 	}
// 	DebugLog("TlsCertCenter start decord conn to %v from %v", host, r.RemoteAddr)
// 	// conn := tls.Server(ws, config)
// 	// io.Copy(conn, conn)
// }}
// server.ServeHTTP(w, r)
// }

// func (t *TLSCertCenter) tlsConfig(host, certFile, keyFile string) (config *tls.Config, err error) {
// 	t.certsLck.RLock()
// 	config = t.loaded[host]
// 	t.certsLck.RUnlock()
// 	if config != nil {
// 		return
// 	}
// 	config = &tls.Config{}
// 	config.NextProtos = append(config.NextProtos, "http/1.1")
// 	config.Certificates = make([]tls.Certificate, 1)
// 	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
// 	return
// }

func (t *TLSCertCenter) cert(w http.ResponseWriter, r *http.Request, host, cert, key string) {
	if len(cert) < 1 || len(key) < 1 {
		w.WriteHeader(500)
		fmt.Fprintf(w, "host config is invalid")
		WarnLog("TlsCertCenter the %v config is missing cert/key", host)
		return
	}
	certBytes, err := ioutil.ReadFile(cert)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "host config is invalid")
		WarnLog("TlsCertCenter the %v config read cert from %v fail with %v", host, cert, err)
		return
	}
	keyBytes, err := ioutil.ReadFile(key)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "host config is invalid")
		WarnLog("TlsCertCenter the %v config read key from %v fail with %v", host, key, err)
		return
	}
	outBytes, _ := json.Marshal(map[string]interface{}{
		"host": host,
		"cert": certBytes,
		"key":  keyBytes,
	})
	w.Write(outBytes)
}

//TLSDecorder provider Decorder to decord conenction by tls cert
type TLSDecorder struct {
	Name      string
	Server    string
	Username  string
	Password  string
	Cert, Key string
	loaded    *tls.Config
	locker    sync.RWMutex
}

//NewTLSDecorder will create new TLSDecorder
func NewTLSDecorder() (decorder *TLSDecorder) {
	decorder = &TLSDecorder{
		locker: sync.RWMutex{},
	}
	return
}

//Decord will decord raw connection by host
func (t *TLSDecorder) Decord(host string, raw net.Conn) (conn net.Conn, err error) {
	t.locker.Lock()
	defer t.locker.Unlock()
	if t.loaded != nil {
		conn = tls.Server(raw, t.loaded)
		return
	}
	config := &tls.Config{}
	config.NextProtos = append(config.NextProtos, "http/1.1")
	config.Certificates = make([]tls.Certificate, 1)
	if len(t.Cert) > 0 && len(t.Key) > 0 {
		config.Certificates[0], err = tls.LoadX509KeyPair(t.Cert, t.Key)
		if err != nil {
			InfoLog("TLSDecorder load X509KeyPair file fail with %v", t.Server, err)
			return
		}
	} else {
		var certData []byte
		certData, err = httpGet(fmt.Sprintf(t.Server, host), t.Username, t.Password)
		if err != nil {
			InfoLog("TLSDecorder send request by %v fail with %v", t.Server, err)
			return
		}
		certInfo := map[string]interface{}{}
		err = json.Unmarshal(certData, &certInfo)
		if err != nil {
			InfoLog("TLSDecorder parse cert info fail with %v by response:\n%v\n", err, string(certData))
			return
		}
		certEncoded, _ := certInfo["cert"].(string)
		cert, _ := base64.StdEncoding.DecodeString(certEncoded)
		keyEncoded, _ := certInfo["key"].(string)
		key, _ := base64.StdEncoding.DecodeString(keyEncoded)
		config.Certificates[0], err = tls.X509KeyPair(cert, key)
		if err != nil {
			InfoLog("TLSDecorder load X509KeyPair fail with %v", err)
			return
		}
	}
	t.loaded = config
	conn = tls.Server(raw, t.loaded)
	return
}
