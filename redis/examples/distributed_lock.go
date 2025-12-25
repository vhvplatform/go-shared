//go:build ignore
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/vhvcorp/go-shared/redis"
)

func main() {
	// Connect to Redis
	client, err := redis.NewClient(redis.Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer client.Close()

	// Create cache
	cache := redis.NewCache(client, redis.CacheConfig{
		DefaultTTL: 5 * time.Minute,
		KeyPrefix:  "lock-example",
	})

	ctx := context.Background()

	// Example 1: Basic Lock Usage
	fmt.Println("=== Example 1: Basic Lock Usage ===")
	
	lock := cache.Lock("resource:1", 10*time.Second)
	
	err = lock.Acquire(ctx, 5*time.Second)
	if err != nil {
		log.Fatalf("Failed to acquire lock: %v", err)
	}
	fmt.Println("Lock acquired")
	
	// Do work while holding lock
	fmt.Println("Processing resource...")
	time.Sleep(2 * time.Second)
	
	err = lock.Release(ctx)
	if err != nil {
		log.Printf("Failed to release lock: %v", err)
	}
	fmt.Println("Lock released\n")

	// Example 2: WithLock Helper
	fmt.Println("=== Example 2: WithLock Helper ===")
	
	err = cache.WithLock(ctx, "resource:2", 10*time.Second, func() error {
		fmt.Println("Lock acquired automatically")
		fmt.Println("Processing resource...")
		time.Sleep(1 * time.Second)
		fmt.Println("Done processing")
		return nil
	})
	if err != nil {
		log.Printf("Error: %v", err)
	}
	fmt.Println("Lock released automatically\n")

	// Example 3: Lock Contention
	fmt.Println("=== Example 3: Lock Contention ===")
	
	var wg sync.WaitGroup
	counter := 0
	
	// Simulate 5 goroutines trying to update a counter
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			err := cache.WithLock(ctx, "counter:lock", 5*time.Second, func() error {
				fmt.Printf("Goroutine %d acquired lock\n", id)
				
				// Read-modify-write
				temp := counter
				time.Sleep(100 * time.Millisecond) // Simulate processing
				counter = temp + 1
				
				fmt.Printf("Goroutine %d updated counter to %d\n", id, counter)
				return nil
			})
			
			if err != nil {
				log.Printf("Goroutine %d error: %v\n", id, err)
			}
		}(i)
	}
	
	wg.Wait()
	fmt.Printf("Final counter value: %d\n\n", counter)

	// Example 4: Lock Extension
	fmt.Println("=== Example 4: Lock Extension ===")
	
	lock = cache.Lock("long:task", 5*time.Second)
	err = lock.Acquire(ctx, 1*time.Second)
	if err != nil {
		log.Fatalf("Failed to acquire lock: %v", err)
	}
	fmt.Println("Lock acquired with 5 second TTL")
	
	// Simulate long-running task
	for i := 1; i <= 3; i++ {
		fmt.Printf("Processing chunk %d...\n", i)
		time.Sleep(2 * time.Second)
		
		// Extend lock before it expires
		err = lock.Extend(ctx, 5*time.Second)
		if err != nil {
			log.Printf("Failed to extend lock: %v", err)
		}
		fmt.Println("Lock extended")
	}
	
	lock.Release(ctx)
	fmt.Println("Task completed, lock released\n")

	// Example 5: Auto-Refresh Lock
	fmt.Println("=== Example 5: Auto-Refresh Lock ===")
	
	lock = cache.Lock("auto:refresh", 3*time.Second)
	err = lock.Acquire(ctx, 1*time.Second)
	if err != nil {
		log.Fatalf("Failed to acquire lock: %v", err)
	}
	fmt.Println("Lock acquired")
	
	// Start auto-refresh in background
	refreshCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Cast to *RedisLock to access RefreshLoop
	redisLock := lock.(*redis.RedisLock)
	errChan := redisLock.RefreshLoop(refreshCtx, 1*time.Second)
	fmt.Println("Auto-refresh started (every 1 second)")
	
	// Do long work
	fmt.Println("Processing long task...")
	time.Sleep(8 * time.Second)
	fmt.Println("Long task completed")
	
	// Stop auto-refresh
	cancel()
	
	// Check for refresh errors
	select {
	case err := <-errChan:
		if err != nil {
			log.Printf("Refresh error: %v", err)
		}
	default:
		fmt.Println("Auto-refresh stopped successfully")
	}
	
	lock.Release(ctx)
	fmt.Println("Lock released\n")

	// Example 6: Lock Timeout Handling
	fmt.Println("=== Example 6: Lock Timeout Handling ===")
	
	// Hold lock in one goroutine
	lock1 := cache.Lock("busy:resource", 10*time.Second)
	err = lock1.Acquire(ctx, 1*time.Second)
	if err != nil {
		log.Fatalf("Failed to acquire lock: %v", err)
	}
	fmt.Println("First lock acquired")
	
	// Try to acquire same lock with short timeout
	go func() {
		lock2 := cache.Lock("busy:resource", 10*time.Second)
		fmt.Println("Trying to acquire busy lock...")
		
		err := lock2.Acquire(ctx, 2*time.Second)
		if err == redis.ErrLockNotAcquired {
			fmt.Println("Could not acquire lock (timeout)")
		} else if err != nil {
			log.Printf("Error: %v", err)
		}
	}()
	
	// Release after some time
	time.Sleep(5 * time.Second)
	lock1.Release(ctx)
	fmt.Println("First lock released\n")

	// Example 7: Distributed Task Processing
	fmt.Println("=== Example 7: Distributed Task Processing ===")
	
	processTask := func(taskID string) error {
		lockKey := fmt.Sprintf("task:%s", taskID)
		
		return cache.WithLock(ctx, lockKey, 30*time.Second, func() error {
			fmt.Printf("Processing task %s...\n", taskID)
			
			// Simulate task processing
			time.Sleep(1 * time.Second)
			
			fmt.Printf("Task %s completed\n", taskID)
			return nil
		})
	}
	
	// Simulate multiple workers trying to process the same task
	var taskWG sync.WaitGroup
	taskID := "task-123"
	
	for i := 1; i <= 3; i++ {
		taskWG.Add(1)
		go func(workerID int) {
			defer taskWG.Done()
			
			fmt.Printf("Worker %d attempting task %s\n", workerID, taskID)
			err := processTask(taskID)
			if err != nil {
				if err == redis.ErrLockNotAcquired {
					fmt.Printf("Worker %d: task already being processed\n", workerID)
				} else {
					log.Printf("Worker %d error: %v\n", workerID, err)
				}
			}
		}(i)
	}
	
	taskWG.Wait()
	fmt.Println("\nAll workers completed")

	// Clean up
	cache.FlushPrefix(ctx)
	fmt.Println("Cleanup complete")
}
