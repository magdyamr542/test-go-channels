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
	// Ex4()
	// Ex5()
	// Ex6()
	// Ex7()
	Ex8()
}

// example of rate limiting.
// the number of requests handled within a given time range can't exceed a given limit
func Ex8() {
	requests := make(chan int)

	requestsLimit := 3 // the max num of requests handled within requestsTime
	requestsTime := time.Second * 1

	quotas := make(chan struct{}) // to handle a request. you must have a quota.

	go func() {
		ticker := time.NewTicker(requestsTime / time.Duration(requestsLimit))
		for range ticker.C {
			fmt.Println("Tick...")
			select {
			case quotas <- struct{}{}:
			default:
				// request are dropped here because there is load on the resources
				fmt.Println("Dropped")
			}
		}
	}()

	go func() {
		for request := range requests {
			<-quotas
			// handle the request
			go func(req int) {
				fmt.Println("request", req)
			}(request)
		}
	}()

	for i := 0; ; i++ {
		requests <- i
	}

}

// fixes the data race introduces in ex6
func Ex7() {
	done := make(chan struct{})
	count := 0
	var mu sync.Mutex
	for i := 0; i < 4; i++ {
		go func(id int) {
			for i := 0; i < 2; i++ {
				fmt.Println(id, "inc count")
				mu.Lock()
				count++
				mu.Unlock()
			}
			done <- struct{}{}
		}(i)
	}
	<-done
	<-done
	<-done
	<-done
	fmt.Println("count", count)
}

// shows a data race where two go routines try to write a variable without coordination
func Ex6() {
	done := make(chan struct{})
	count := 0
	for i := 0; i < 4; i++ {
		go func(id int) {
			for i := 0; i < 2; i++ {
				fmt.Println(id, "inc count")
				count++
			}
			done <- struct{}{}
		}(i)
	}
	<-done
	<-done
	<-done
	<-done
	fmt.Println("count", count)
}

// implements a fan in. multiple channels produce. one channel outputs all produces data
func Ex5() {

	// create a data generator
	generator := func(everySecond int, id int) <-chan string {
		c := make(chan string)
		count := 0
		go func() {
			defer close(c)
			for {
				time.Sleep(time.Duration(everySecond) * time.Second)
				c <- fmt.Sprintf("id=%d: %d", id, count)
				count += 1
				if count == 10 {
					return
				}

			}

		}()
		return c
	}

	fanIn := func(sources ...<-chan string) <-chan string {
		// this needs to be closed for readers not to block
		// it should be closed after all sources are done sending data
		output := make(chan string)
		var wg sync.WaitGroup
		for _, source := range sources {
			wg.Add(1)
			go func(source <-chan string) {
				for value := range source {
					output <- value
				}
				wg.Done()
			}(source)
		}

		go func() {
			wg.Wait()
			close(output)
		}()
		return output
	}

	producer1 := generator(1, 1)
	producer2 := generator(1, 2)

	fannedIn := fanIn(producer1, producer2)

	for readValue := range fannedIn {
		fmt.Println(readValue)
	}
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
