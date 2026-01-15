package main

import (
	"context"
	"log"
	"sync"
	"time"

	clientv3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

/// Etcd is a distributed key-value store that provides a reliable way to store data across a cluster of machines.
// It is often used for configuration management, service discovery, and coordination of distributed systems.
func main() {
	// 1) Create a client to connect to the Etcd server
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create Etcd client: %v", err)
	}
	defer client.Close()

	// 2) Perform concurrent read and write operations
	var wg sync.WaitGroup
	lockName := "my-distributed-lock"

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// 2) create session 
			session, err := concurrency.NewSession(client, 
				concurrency.WithTTL(10),
			)
			if err != nil {
				log.Printf("Goroutine %d: Failed to create session: %v", id, err)
				return
			}
			defer session.Close()

			// 3) create Mutex based on session
			mutex := concurrency.NewMutex(session, "/locks/"+lockName)

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			// 5) add lock
			if err := mutex.Lock(ctx); err != nil {
				log.Printf("Goroutine %d: Failed to acquire lock: %v", id, err)
				return
			}
			log.Printf("Goroutine %d: Acquired lock", id)
			 
			// Simulate some work with the lock held
			time.Sleep(2 * time.Second)

			// 6) release lock
			if err := mutex.Unlock(ctx); err != nil {
				log.Printf("Goroutine %d: Failed to release lock: %v", id, err)
				return
			}
			log.Printf("Goroutine %d: Released lock", id)

		}(i)
	}

	wg.Wait()
	log.Println("All goroutines have completed.")
}