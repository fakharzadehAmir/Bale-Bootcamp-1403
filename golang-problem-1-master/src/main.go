package main

import (
	"time"
)

func Solution(d time.Duration, message string, ch ...chan string) (numberOfAccesses int) {
	// if we wanna use this function in concurrent goroutines,
	// increament of `numberOfAccesses` must be atomic. (using sync.Mutex)
	// mtx := &sync.Mutex{}
	timer := time.NewTimer(d * time.Second)
	for {
		select {
		case <-timer.C:
			return numberOfAccesses
		default:
			for _, channel := range ch {
				if len(channel) == 0 {
					select {
					case <-timer.C:
						return numberOfAccesses
					case channel <- message:
						// mtx.Lock()
						numberOfAccesses++
						// mtx.Unlock()
					default:
					}

				}
			}

		}
	}
}
