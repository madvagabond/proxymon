package proxymon

import (
	"http"
	"strings"
	socks_client "h12.io/socks"

	"errors"
	
)


type Protocol = string 

const socks4 Protocol = "socks4"
const socks4a Protocol = "socks4a"
const socks5 Protocol = "socks5"

const http Protocol = "http"
const https Protocol = "https"



type Client interface {
	NewClient( Proxy) *http.Client
	Protocol() Protocol
}


type Proxy struct {
	Protocol string
	Username string
	Password string
	Host string 
}





type Socks struct {
	Version string 
	Username string
	Password string
	Host string 
}

/* Uses Version field if it's empty returns socks4 */



func matchProtocol(v string) (Protocol, error) {
	
	version := strings.ToLower(v)
	
	if strings.Contains(version, "socks4a") {
		return socks4a
	}
	
	if strings.Contains(version, "socks4") {
		return socks4
	}


	if strings.Contains(version, "socks5") {
		return socks5
	}


	if strings.Contains(version, "https") {
		return https
	}

	if strings.Contains(version, "http") {
		return http 
	} 



	return errors.New("Unknown Protocol")  
}



func (s *Socks) get_protocol() Protocol {


	version := strings.ToLower(s.Version)
	
	if strings.Contains(version, "4a") {
		return socks4a
	}
	
	if strings.Contains(version, "4") {
		return socks4
	}


	if strings.Contains(version, "5") {
		return socks5
	}





	return socks4
}




func before(value string, a string) string {
    // Get substring before a string.
    pos := strings.Index(value, a)
    if pos == -1 {
        return ""
    }
    return value[0:pos]
}


func after(value, a string) string {
	pos := strings.Index(value, a)
	l := len(a)
	
	if pos == -1 {return ""}
	

	return value[pos+l:]
}






func userPass(s string) (string, string) {


	
	s1 := after(s, "://")
	s2 := before(s1, "@")


	if s2 == "" {return "", ""}
	
	toks := strings.Split(s2, ":")

	l := len(toks)


	if l >= 2 {
		return toks[0], toks[1]	
	}


	if l == 1 {
		return toks[0], ""
	}

	return "", ""	
}




func getHost(s string) string {

	host := after(s, "@")
	
	if host == "" {
		return s
	}

	return host 
	
} 

func (s *Socks) FromUri(s string) {
	prefix := before(s, "://")
	proto, e := matchProtocol(prefix)
	
	user, pass := userPass(s)
	host := after(s, "@")

	if e != nil {proto = socks4 }

	switch prefix {
	case socks4:
		Socks4
	}
}


func (s *Socks) ToUri() string {

	var prefix string

	protocol := s.Protocol() 

	if protocol == socks4 {
		prefix = "socks4://"
	}

	if protocol == socks5 {
		prefix = "socks5://"
	}


	if protocol == socks4a {
		prefix = "socks4a://"
	}


	uri := prefix + s.User + ":" s.Password + "@" + s.Host
	return uri  

	
}

func (s *Socks) DialContext(ctx context.Context, net, addr string) (net.Conn, error) {

	uri := s.proxyUri()

	dialer := socks_client.Dial(uri)

	return dialer(net, addr) 
}


func (s *Socks) NewClient() *http.Client {
	
	return &http.Client{
		Transport: &http.Transport{
			DialContext: dialCtx,
		}, 
	}
	
}


