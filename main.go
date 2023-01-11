package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	// Ex1()
	// Ex2()
	// Ex3()
	Ex4()
}

// implementing a ticker
func Ex4() {

	ticker := func(every time.Duration, stop <-chan struct{}) chan struct{} {

		result := make(chan struct{})

		go func() {
			defer func() {
				close(result) // close result to avoid deadlock from other goroutines which are reading from result
			}()

			for {
				time.Sleep(every)
				select {
				case result <- struct{}{}:
				case <-stop:
					fmt.Println("Stopping...")
					return
				default:
					fmt.Println("Default case didn't send a tick signal to the channel")

				}
			}
		}()
		return result
	}

	stop := make(chan struct{})
	result := ticker(time.Second, stop)
	t := time.Now()
	for range result {
		since := time.Since(t)
		fmt.Println("Ticking", since)
		if since.Seconds() > 4 {
			fmt.Println("Will stop")
			close(stop)
		}
	}

	fmt.Println("Done")
}

// first response wins
// this implements a try-send mechanism
func Ex3() {

	data := make(chan string)

	for i := 0; i < 4; i++ {
		go func(id int) {
			sleeping := 3
			fmt.Printf("Worker %d sleeping for %d seconds\n", id, sleeping)
			time.Sleep(time.Duration(sleeping) * time.Second)
			// will try to send. if it didn't work. will just exec the default
			select {
			case data <- strconv.Itoa(id):
			default:
				fmt.Printf("Worker %d didn't send\n", id)
			}
		}(i)
	}
	fmt.Printf("Data from worker is %s\n", <-data)
}

// workers get jobs and then notify on finish
func Ex2() {
	nWorkers := 4
	wg := sync.WaitGroup{}
	wg.Add(nWorkers)
	work := func(id int) {
		defer wg.Done()
		sleeping := rand.Intn(4)
		fmt.Printf("Worker %d started and working for %d seconds\n", id, sleeping)
		time.Sleep(time.Duration(sleeping) * time.Second)
	}
	for i := 0; i < nWorkers; i++ {
		go work(i)
	}
	fmt.Printf("Waiting for all workers to finish\n")
	wg.Wait()
	fmt.Printf("Work is done\n")
}

// simple example. notify when done
func Ex1() {
	fmt.Println("Starting...")
	done := make(chan int)
	go func() {
		sleeping := rand.Intn(4)
		fmt.Printf("Sleeping for %d seconds\n", sleeping)
		time.Sleep(time.Duration(sleeping) * time.Second)
		done <- sleeping
	}()
	fmt.Printf("Ending after waiting for %d seconds", <-done)
}
