package speedtest


import (
	"encoding/xml"
	"math"
	"time"

	"net/http"


	"strconv"

)



type client_config struct {
	XMLName xml.Name `xml:client`
	Lon float64 `xml:lon`
	Lat float64 `xml:lat`
}




type Client struct {
	http_client *http.Client
	server Server
}


type Result struct {
	UploadSpeed float64
	DownloadSpeed float64
	PingLatency time.Duration
}


func distance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	radius := 6378.137

	a1 := lat1 * math.Pi / 180.0
	b1 := lon1 * math.Pi / 180.0
	a2 := lat2 * math.Pi / 180.0
	b2 := lon2 * math.Pi / 180.0

	x := math.Sin(a1)*math.Sin(a2) + math.Cos(a1)*math.Cos(a2)*math.Cos(b2-b1)
	return radius * math.Acos(x)
}





func (cli client_config) distance(srv *Server) float64 {

	f64 := func (str string) float64 {
		
		f, _ :=  strconv.ParseFloat(str, 64)
		return f
	}

	
	return distance(cli.Lat, cli.Lon, f64(srv.Lat), f64(srv.Lon))
} 



func (cli client_config) closest(srvs []Server) Server {


	closest := srvs[0]
	

	
	for _, v := range srvs {
		dist := cli.distance(&v)
		mdist := cli.distance(&closest)

		if dist < mdist {
			closest = v 
		}	
	}


	return closest 
	
}




func NewClient(http_client *http.Client) (Client, error) {
	servers, e := getServers(http_client)
	config, e :=  getClientConf(http_client)

	
	var c Client
	if e != nil {return c, e}

	server := config.closest(servers)

	return Client{http_client, server}, nil
}


func (cli *Client) DownloadTest() (float64, error) {

	dl_f := func () (int, error) {
		return downloadReq(cli.http_client, cli.server.URL)
	}

	return checkSpeed(dl_f)
}



func (cli *Client) UploadTest() (float64, error) {
	ul_f := func() (int, error) {
		return uploadReq(cli.http_client, cli.server.URL)
	}

	return checkSpeed(ul_f)
}



func (cli *Client) PingTest() (time.Duration, error) {
	return pingTest(cli.http_client, cli.server.URL)
}


func (cli *Client) SpeedTest() (Result, error) {
	dl, e := cli.DownloadTest()
	ul, e := cli.UploadTest()
	p, e := cli.PingTest()


	
	if e != nil {
		var r Result
		return r, e
	}

	
	res := Result{DownloadSpeed: dl, UploadSpeed: ul, PingLatency: p}

	return res, e 
}
