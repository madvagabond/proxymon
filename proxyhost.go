package proxymon

import (
	//"net/http"
	"strings"
	"errors"	
)


type Proxy struct {
	Protocol string
	Username string
	Password string
	Host string 
}


func parseProto(v string) (string, error) {
	tok := strings.ToLower( before(v, "://") )

	switch tok {
	case "socks":
		return "socks4", nil

	case "socks5":
		return tok, nil

	case "socks4a":
		return tok, nil

	case "http":
		return tok, nil

	case "https":
		return tok, nil

	default:
		return "", errors.New("invalid proxy protocol scheme") 
	}
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



func parseAuth(s string) (string, string) {


	
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




func parseHost(s string) string {

	host := after(s, "@")
	
	if host == "" {
		return s
	}

	return after(host, "://")  
	
} 



func ParseProxyString(s string) (Proxy, error) {
	proto, e := parseProto(s)
	
	if e != nil {
		var p Proxy
		return p, e
	}

	/*
	   used after :// twice is a bit inefficient for http
	   
	*/
	user, pass := parseAuth(s)
	host := parseHost(s)

	p := Proxy{
		Protocol: proto,
		Username: user,
		Password: pass,
		Host: host,
	}


	return p, nil
}


func (p *Proxy) ToUri() string {

	return p.Protocol + "://" + p.Username + ":" + p.Password + "@" + p.Host 
}


