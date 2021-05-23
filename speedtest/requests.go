package speedtest

import (
	"net/http"
	"io"
	"encoding/xml"
	"io/ioutil"
	"time"
	"strings"
	"strconv"

	"net/url"
	"math"
)




var servers_uri = "https://www.speedtest.net/speedtest-servers-static.php"






type servers struct {
	Servers []Server `xml:"servers>server"`
}

type Server struct {
	URL      string `xml:"url,attr"`
	Lat      string `xml:"lat,attr"`
	Lon      string `xml:"lon,attr"`
	Name     string `xml:"name,attr"`
	Country  string `xml:"country,attr"`
	Sponsor  string `xml:"sponsor,attr"`
	ID       string `xml:"id,attr"`
	URL2     string `xml:"url2,attr"`
	Host     string `xml:"host,attr"`
}


func getServers(cli *http.Client) ([]Server, error) {
	rep, err := cli.Get(servers_uri)
	data, err := io.ReadAll(rep.Body) 

	var servers servers

	if err != nil {
		return []Server{}, err
	}

	
	e := xml.Unmarshal(data, &servers)
	
	if e != nil {return []Server{}, e}

	return servers.Servers, nil 
		
}




func getClientConf(cli *http.Client) (client_config, error) {
	resp, err := http.Get("http://speedtest.net/speedtest-config.php")

	if err != nil {
		var c client_config
		return c, nil  
	}


	var c client_config
	body, e := io.ReadAll(resp.Body )

	if err != nil {return c, e}
	
	err = xml.Unmarshal(body, &c)
	return c, e
	
} 







func downloadReq(cli *http.Client, url string) (int, error) {

	size := 1500

	dlURL := strings.Split(url, "/upload")[0]
	uri  := dlURL + "/random" + strconv.Itoa(size) + "x" + strconv.Itoa(size) + ".jpg"


	
	resp, err := cli.Get(uri)
	if err != nil {return 0, err}

	buf, e := ioutil.ReadAll(resp.Body)
	
	resp.Body.Close()

	return len(buf), e
}





func pingTest(cli *http.Client, url string) (time.Duration, error) {


	ep := strings.Split(url, "/upload.php")[0] + "/latency.txt"
 

	start := time.Now()
	
	_, err := cli.Get(ep)
	end := time.Now() 
	
	if err != nil {
		return 0, err
	}


	return end.Sub(start), nil
}






func checkSpeed(call func() (int, error)) (float64, error ){
	start := time.Now()
	size, e := call()
	end := time.Now()
	
	if e != nil {return 0.0, e}
	t := end.Sub(start)

	megs := float64(size) / math.Pow10(6)
	rate := megs / t.Seconds()

	return rate, nil 
}



func uploadReq(cli *http.Client, uri string) (int, error) {
	size := 1000
	v := url.Values{}

	
	v.Add("content", strings.Repeat("0123456789", size*100-51))

	resp, err := cli.PostForm(uri, v)

	if err != nil {
		return 0, err
	}
	
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)

	written := size * 10 * 100-51
	return written, nil
} 

