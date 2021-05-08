package proxymon
import (
	"time"
	"sync"

)



/* Add Availability Tracker Metric Later*/

type Metrics struct {
	Upload float
	Download float
	Ping float

	SuccessfulPings int 
	FailedPings int 
	LastSeen time.Time
}







func (m *Metrics) Uptime() float {
	total := m.SuccessfulPings + m.FailedPings
	return m.SuccessfulPings / total
} 



func (m *Metrics) Merge(res Result) {

	t := time.Now()
	m.Upload = res.UploadSpeed
	m.Download = res.DownloadSpeed
	m.Ping = res.PingLatency
	
	m.SuccessfulPings + 1
	m.LastSeen
}

type Node struct {
	Proxy
	Metric
}
 


type Config struct {
	PingInterval time.Duration
	DeadExpiry time.Duration
	BandwidthTestInterval time.Duration
	ConcurrentBandwidthTests int

}





type EventDelegate interface {
	func HandleAdd(n Node)
	func HandleDead(n Node)
	func HandleRemove(n Node)
	func HandleUpdate(n Node)
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
	
	tester, e := NewSpeedTester(p)
	
	if e != nil {
		var res speedtest.Result
		return res, e
	}


	return tester.SpeedTest()

}







func makeNode(p Proxy, res speedtest.Result) Node {

	t := time.Now()


	return Node{
		Ping: res.PingLatency,
		Download: res.DownloadSpeed,
		UploadSpeed: res.UploadSpeed,
		LastSeen: t,
		SuccessfulPings: 1,
		FailedPings: 0, 
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
	go m.delegate.HandleAdd()

	
}


func (m *Monitor) Add(p Proxy) {

	m.lock.RLock()
	_, e := up[p]
	_, e1 := down[p]

	m.lock.RUnlock()


	
	
	if e || e1 {
		return 
	}


	res, e := testSpeed(p)

	if e != nil {
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


	<- m.bw_permits
	res, e := testSpeed(n)

	bw_permits <- 1
	
	
	if e != nil {
		m.markDead(n)
		return
	}

	m.addOrUpdate(n, res)
}




func (m *Monitor) pingSuccess(p Proxy, t time.Time) {
	defer m.lock.Unlock()
	m.lock.Lock()

	a, e := m.up[p]
	d, e1 := m.down[p]


	if e {

		a.SuccessfulPings += 1 
		a.LastSeen = t 
		m.up[p] = a
	}


	if e1 {
		a.SuccessfulPings += 1 
		d.LastSeen = t
		
		delete(m.down, p)
		m.up[p] = d
	}


}


func (m *Monitor) pingNode(n Node) {


	tester, e := NewSpeedTester(p)
	
	if e != nil {
		m.markDead(n)
		return 
	}

	
	l, e := tester.PingTest()
	t := time.Now()
	
	if e != nil {
		m.markDead()
		return 
	}

	m.pingSuccess(n)
} 
