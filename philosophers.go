package main

import (
	"fmt"
	"sync"
	"time"
)

// Chopstick represents a chopstick
type Chopstick struct {
	id     int
	access sync.Mutex
}

// Philosopher represents a philosopher
type Philosopher struct {
	id             int
	leftChopStick  *Chopstick
	rightChopStick *Chopstick
}

// Host manages permissions
type Host struct {
	currentServingCount int
	access              sync.Mutex
}

func getChopsticks(count int) []*Chopstick {
	chopsticks := make([]*Chopstick, count)
	for i := 0; i < count; i++ {
		chopsticks[i] = &Chopstick{id: i}
	}

	return chopsticks
}

func getPhilosophers(count int) []*Philosopher {
	philosophers := make([]*Philosopher, count)
	for i := 0; i < count; i++ {
		philosophers[i] = &Philosopher{id: i}
	}

	return philosophers
}

func arrangeChopsticks(philophers []*Philosopher, chopSticks []*Chopstick, count int) {
	for i, philosopher := range philophers {
		philosopher.leftChopStick = chopSticks[i]
		philosopher.rightChopStick = chopSticks[(i+1)%count]
	}
}

// Lock locks access to the chopstick
func (c *Chopstick) Lock() {
	c.access.Lock()
}

// Unlock unlocks access to the chopstick
func (c *Chopstick) Unlock() {
	c.access.Unlock()
}

func (philosopher *Philosopher) lockChopsticks() {
	philosopher.leftChopStick.Lock()
	philosopher.rightChopStick.Lock()
}

func (philosopher *Philosopher) unlockChopsticks() {
	philosopher.leftChopStick.Unlock()
	philosopher.rightChopStick.Unlock()
}

func (philosopher *Philosopher) eat(eatingStart *chan string, eatingEnd *chan string) {
	philosopher.lockChopsticks()

	*eatingStart <- "eating started"

	fmt.Println("starting to eat ", philosopher.id)
	time.Sleep(100)
	fmt.Println("finishing eating ", philosopher.id)

	philosopher.unlockChopsticks()

	*eatingEnd <- "eating ended"
}

func (host *Host) processRequest(servingAvailable *chan string, eatingStart *chan string) {
	if host.currentServingCount < 2 {
		host.access.Lock()
		if host.currentServingCount < 2 {
			*servingAvailable <- "serving available"
			<-*eatingStart
			host.currentServingCount++
			host.access.Unlock()

			return
		}

		host.access.Unlock()
	}

	host.processRequest(servingAvailable, eatingStart)
}

func (host *Host) acceptRequest(philosopher *Philosopher, requestID int) *chan string {
	eatingStart := make(chan string)
	eatingEnd := make(chan string)
	servingAvailable := make(chan string)
	requestClosure := make(chan string)

	go func() {
		<-servingAvailable
		philosopher.eat(&eatingStart, &eatingEnd)
	}()

	go func() {
		host.processRequest(&servingAvailable, &eatingStart)
	}()

	go func() {
		<-eatingEnd
		host.access.Lock()
		host.currentServingCount--
		host.access.Unlock()
		requestClosure <- "request closed"
	}()

	return &requestClosure
}

func main() {
	count := 5
	servingCount := 3
	chopsticks := getChopsticks(count)
	philosophers := getPhilosophers(count)
	host := new(Host)
	arrangeChopsticks(philosophers, chopsticks, count)

	requestClosures := []*chan string{}

	for _, ph := range philosophers {
		for i := 0; i < servingCount; i++ {
			requestClosure := host.acceptRequest(ph, ph.id*3+i)
			requestClosures = append(requestClosures, requestClosure)
		}
	}

	for _, rc := range requestClosures {
		<-*rc
	}

	fmt.Println("all done eating!")
}
