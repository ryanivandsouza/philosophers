package main

import (
	"fmt"
	"sync"
	"time"
)

// Chopstick represents a chopstick
type Chopstick struct {
	sync.Mutex
}

// Philosopher represents a philosopher
type Philosopher struct {
	id             int
	leftChopStick  *Chopstick
	rightChopStick *Chopstick
	eatingEnded    chan string
}

// Host manages servings available
type Host struct {
	servingAvailable chan string
}

func getChopsticks(count int) []*Chopstick {
	chopsticks := make([]*Chopstick, count)
	for i := 0; i < count; i++ {
		chopsticks[i] = &Chopstick{}
	}

	return chopsticks
}

func getPhilosophers(count int) []*Philosopher {
	philosophers := make([]*Philosopher, count)
	for i := 0; i < count; i++ {
		philosophers[i] = &Philosopher{id: i, eatingEnded: make(chan string)}
	}

	return philosophers
}

func setTable(philophers []*Philosopher, chopSticks []*Chopstick, count int) {
	for i, philosopher := range philophers {
		philosopher.leftChopStick = chopSticks[i]
		philosopher.rightChopStick = chopSticks[(i+1)%count]
	}
}

func (philosopher *Philosopher) lockChopsticks() {
	philosopher.leftChopStick.Lock()
	philosopher.rightChopStick.Lock()
}

func (philosopher *Philosopher) unlockChopsticks() {
	philosopher.leftChopStick.Unlock()
	philosopher.rightChopStick.Unlock()
}

func (philosopher *Philosopher) eat() {
	philosopher.lockChopsticks()

	fmt.Println("starting to eat ", philosopher.id)
	time.Sleep(100)
	fmt.Println("finishing eating ", philosopher.id)

	philosopher.unlockChopsticks()

	philosopher.eatingEnded <- "eating ended"
}

func (host *Host) acceptRequest(philosopher *Philosopher, requestID int, wg *sync.WaitGroup) {
	go func() {
		<-host.servingAvailable
		philosopher.eat()
	}()

	go func() {
		<-philosopher.eatingEnded
		host.servingAvailable <- "serving available"
		wg.Done()
	}()
}

// MaxConcurrentServings specifies maximum number of servings allowed at time
var MaxConcurrentServings = 2

// MaxNumberOfServingsPerPerson specifies maximum number of servings per person
var MaxNumberOfServingsPerPerson = 3

// PhilosopherCount specifies number of philosophers. Must be greater than 1 (>1)
var PhilosopherCount = 5

func main() {
	philosophers := getPhilosophers(PhilosopherCount)
	chopsticks := getChopsticks(PhilosopherCount)
	setTable(philosophers, chopsticks, PhilosopherCount)

	host := &Host{servingAvailable: make(chan string, MaxConcurrentServings)}

	go func() {
		host.servingAvailable <- "serving available"
		host.servingAvailable <- "serving available"
	}()

	var wg sync.WaitGroup

	for i := 0; i < MaxNumberOfServingsPerPerson; i++ {
		for _, ph := range philosophers {
			wg.Add(1)
			go host.acceptRequest(ph, ph.id*3+i, &wg)
		}
	}

	wg.Wait()

	fmt.Println("all done eating!")
}
