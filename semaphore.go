package proxymon

type semaphore struct {
	c chan int 
}

func (s *semaphore) acquire() {
	<- s.c
}


func (s *semaphore) release() {
	s.c <- 1 
}

func make_semaphore(size int) semaphore {
	c := make(chan int, size)
	return semaphore{c}
}
