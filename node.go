package proxymon

import (
	"time"
	"github.com/xnukernpoll/proxymon/speedtest"

)



/* Add Availability Tracker Metric Later*/

type Metrics struct {
	Upload float64
	Download float64
	Ping time.Duration

	SuccessfulPings int 
	FailedPings int 
	LastSeen time.Time
}







func (m *Metrics) Uptime() float64 {
	total := m.SuccessfulPings + m.FailedPings
	return float64(m.SuccessfulPings) / float64(total)
} 



func (m *Metrics) Merge(res speedtest.Result) {

	t := time.Now()
	m.Upload = res.UploadSpeed
	m.Download = res.DownloadSpeed
	m.Ping = res.PingLatency
	
	m.SuccessfulPings += 1
	m.LastSeen = t
}

type Node struct {
	Proxy
	Metrics
}
 








func makeNode(p Proxy, res speedtest.Result) Node {

	t := time.Now()


	return Node{
		p,

		Metrics{
		Ping: res.PingLatency,
		Download: res.DownloadSpeed,
		Upload: res.UploadSpeed,
		LastSeen: t,
		SuccessfulPings: 1,
		FailedPings: 0,
	},
	}

	
}







