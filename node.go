package proxymon

import (
	"time"
	"sync"
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
 


type Config struct {
	PingInterval time.Duration
	DeadExpiry time.Duration
	BandwidthTestInterval time.Duration
	ConcurrentBandwidthTests int

}





type Delegate interface {

	HandleAdd(n Node)
	HandleDead(n Node)
	HandleRemove(n Node)
	HandleUpdate(n Node)
}






type Monitor struct {

	lock sync.RWMutex
	bw_permits semaphore
	up map[Proxy]Node
	down map[Proxy]Node

	delegate Delegate
	config Config 
}


func testSpeed(p Proxy) (speedtest.Result, error) {
	
	tester, e := newSpeedTester(p)
	
	if e != nil {
		var res speedtest.Result
		return res, e
	}


	return tester.SpeedTest()

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

func (m *Monitor) addOrUpdate(proxy Proxy, result speedtest.Result) {
	defer m.lock.Unlock()
	m.lock.Lock()

	
	n, e := m.up[proxy]

	if e {
		n.Merge(result)
		m.up[proxy] = n
		go m.delegate.HandleUpdate(n)
		return 
	}


	n1, e1 := m.down[proxy]

	if e1 {
		n1.Merge(result)
		delete(m.down, proxy)
		m.up[proxy] = n1

		
		go m.delegate.HandleAdd(n)
		return 
	}



	m.up[proxy] = makeNode(proxy, result)
	go m.delegate.HandleAdd(n)

	
}


func (m *Monitor) Add(p Proxy) {

	m.lock.RLock()
	_, e := m.up[p]
	_, e1 := m.down[p]

	m.lock.RUnlock()


	
	
	if e || e1 {
		return 
	}


	res, err := testSpeed(p)

	if err != nil {
		return
	}

	m.addOrUpdate(p, res)
	//fix speed test to retry server lists.
	
}



func (m *Monitor) Nodes() []Node {

	defer m.lock.RUnlock()

	m.lock.RLock()
	nodes := []Node{}

	
	for _, v := range m.up {
		nodes = append(nodes, v)
	}


	return nodes 
}




func (m *Monitor) markDead(p Proxy) {

	defer m.lock.Unlock() 
	m.lock.Lock() 

	a, e := m.up[p]
	d, e1 := m.down[p]

	if e {
		delete(m.up, p)
		a.FailedPings = a.FailedPings + 1
		m.down[p] = a
		go m.delegate.HandleDead(a)
		return
	}


	if e1 {
		d.FailedPings = d.FailedPings + 1
		m.down[p] = d

		
		go m.delegate.HandleDead(d)
		return 
	}

	
} 


func (m *Monitor) speedtest(n Node) {


	m.bw_permits.acquire()
	res, e := testSpeed(n.Proxy)

	m.bw_permits.release()
	
	
	if e != nil {
		m.markDead(n.Proxy)
		return
	}

	m.addOrUpdate(n.Proxy, res)
}




func (m *Monitor) pingSuccess(p Proxy, t time.Time, latency time.Duration) {
	defer m.lock.Unlock()
	m.lock.Lock()

	a, e := m.up[p]
	d, e1 := m.down[p]


	if e {

		a.SuccessfulPings += 1 
		a.LastSeen = t
		a.Ping = latency
		m.up[p] = a
	}


	if e1 {
		d.SuccessfulPings += 1 
		d.LastSeen = t
		d.Ping = latency
		delete(m.down, p)
		m.up[p] = d
	}


}


func (m *Monitor) pingNode(n Node) {


	tester, e := newSpeedTester(n.Proxy)
	
	if e != nil {
		m.markDead(n.Proxy)
		return 
	}

	
	l, e := tester.PingTest()
	t := time.Now()
	
	if e != nil {
		m.markDead(n.Proxy)
		return 
	}

	m.pingSuccess(n.Proxy, t, l)
} 
