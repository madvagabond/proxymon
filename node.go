package proxymon




/* Add Availability Tracker Metric Later*/



type Metrics struct {
	Upload float
	Download float
	Ping float
}



type Node struct {
	Proxy
	Metrics
}



