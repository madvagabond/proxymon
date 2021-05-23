package proxymon

import (
	"time"
	"sync"
	"github.com/xnukernpoll/proxymon/speedtest"
	"log"

)




type Config struct {
	PingInterval time.Duration
	DeadExpiry time.Duration
	BandwidthTestInterval time.Duration
	ConcurrencyLevel int 
}


type Monitor struct {

	lock sync.RWMutex
	bw_permits semaphore
	p_permits semaphore 
	up map[Proxy]Node
	down map[Proxy]Node

	delegate Delegate
	config Config

	bandwidth_timer *time.Timer
	ping_timer *time.Timer
	gc_timer *time.Timer
}






func NewMonitor(c Config, d Delegate) Monitor {
	lock := sync.RWMutex{}
	bw_permits := make_semaphore(c.ConcurrencyLevel)
	p_permits := make_semaphore(c.ConcurrencyLevel)
	bandwidth_timer := time.NewTimer(c.BandwidthTestInterval)
	ping_timer := time.NewTimer(c.PingInterval)
	gc_timer := time.NewTimer(c.DeadExpiry)

	up := make(map[Proxy]Node)
	down := make(map[Proxy]Node)

	
	return Monitor {
		lock,
		bw_permits,
		p_permits,
		up,
		down,
		d,
		c, 
		bandwidth_timer,
		ping_timer,
		gc_timer,
	
	}


	
}


type Delegate interface {

	HandleAdd(n Node)
	HandleDead(n Node)
	HandleRemove(n Node)
	HandleUpdate(n Node)
}


func testSpeed(p Proxy) (speedtest.Result, error) {
	
	tester, e := newSpeedTester(p)
	
	if e != nil {
		var res speedtest.Result
		return res, e
	}


	return tester.SpeedTest()

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
		log.Println("Unable to add proxy because speed test failed")
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







func (m *Monitor) allNodes() []Node {
	var nodes []Node
	
	for _, v := range m.up {
		nodes = append(nodes, v)
	}


	for _, v := range m.down {
		nodes = append(nodes, v)
	}
	return nodes
}



func (m *Monitor) ping_proc() {

	for {
		<- m.ping_timer.C
		for _, v := range  m.allNodes() {
			go m.pingNode(v)
		}

		m.ping_timer.Reset(m.config.PingInterval)
	}
}



func (m *Monitor) bw_proc() {
	for {
		<- m.bandwidth_timer.C
		for _, v := range m.allNodes() {
			go m.speedtest(v)
		}

		m.bandwidth_timer.Reset(m.config.BandwidthTestInterval)

	}

}


