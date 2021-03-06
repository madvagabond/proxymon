package proxymon

import (
	"net/http"
	socks_client "h12.io/socks"
	"strings"
	"net/url"
	"errors"
	"net"
	"context"
	"github.com/xnukernpoll/proxymon/speedtest"
	"log"
)




func httpProxyClient(p Proxy) (*http.Client, error) {
	proxy_s := p.ToUri() 
	p_url, err := url.Parse(proxy_s)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	trans := &http.Transport{ Proxy: http.ProxyURL(p_url) }

	client := http.Client{
		Transport: trans, 
 	}


	return &client, nil 
	
} 


type socksSession struct {
	proxy Proxy
}






type dialResult struct {
	conn net.Conn
	error error
}



func ctxSelector(ctx context.Context, d func() (net.Conn, error) ) (net.Conn, error) {

	res_ch := make(chan dialResult)

	go func() {
		conn, err := d()
		res := dialResult{conn, err}
		res_ch <- res 
	}()

	select {
	case <- ctx.Done():
		var c net.Conn
		return c, errors.New("Context expired before connection could be established.") 


	case res := <- res_ch: 
		return res.conn, res.error
	} 
	
} 


func (s *socksSession) DialContext(ctx context.Context, net, addr string) (net.Conn, error) {	
	uri := s.proxy.ToUri() 
	dialer := socks_client.Dial(uri)
	return dialer(net, addr)
}




func newSocksClient(p Proxy) (*http.Client) {
	session := socksSession{p}
	transport := http.Transport{ DialContext: session.DialContext }
	
	return &http.Client{
		Transport: &transport,
	} 
	
}




func newClient(p Proxy) (*http.Client, error)  {
	if strings.Contains(p.Protocol, "http") {
		return httpProxyClient(p)
	}


	if strings.Contains(p.Protocol, "socks") {
		return newSocksClient(p), nil 
	}


	e := errors.New("Unsupported Protocol")
	var c http.Client
	
	return &c, e 
}



func newSpeedTester(p Proxy) (speedtest.Client, error) {
	c, e := newClient(p)
	if e != nil {
		var cli speedtest.Client
		return cli,e
	}
	
	return speedtest.NewClient(c)
}


