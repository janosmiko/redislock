package redislock_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

func Example() {
	// Connect to redis.
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	defer client.Close()

	// Create a new lock client.
	locker := redislock.New(client)

	ctx := context.Background()

	// Try to obtain lock.
	lock, err := locker.Obtain(ctx, "my-key", 100*time.Millisecond, nil)
	if err == redislock.ErrNotObtained {
		fmt.Println("Could not obtain lock!")
	} else if err != nil {
		log.Fatalln(err)
	}

	// Don't forget to defer Release.
	defer lock.Release(ctx)
	fmt.Println("I have a lock!")

	// Sleep and check the remaining TTL.
	time.Sleep(50 * time.Millisecond)
	if ttl, err := lock.TTL(ctx); err != nil {
		log.Fatalln(err)
	} else if ttl > 0 {
		fmt.Println("Yay, I still have my lock!")
	}

	// Extend my lock.
	if err := lock.Refresh(ctx, 100*time.Millisecond, nil); err != nil {
		log.Fatalln(err)
	}

	// Sleep a little longer, then check.
	time.Sleep(100 * time.Millisecond)
	if ttl, err := lock.TTL(ctx); err != nil {
		log.Fatalln(err)
	} else if ttl == 0 {
		fmt.Println("Now, my lock has expired!")
	}

	// Output:
	// I have a lock!
	// Yay, I still have my lock!
	// Now, my lock has expired!
}

func ExampleClient_Obtain_With_Secret() {
	// Connect to redis.
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	defer client.Close()

	// Create a new lock client.
	locker := redislock.New(client)

	ctx := context.Background()

	opts := &redislock.Options{
		Secret: "test",
	}
	// Try to obtain lock.
	lock, err := locker.Obtain(ctx, "my-key", 150*time.Millisecond, opts)
	if err == redislock.ErrNotObtained {
		fmt.Println("Could not obtain lock!")
	} else if err != nil {
		log.Fatalln(err)
	}

	// Don't forget to defer Release.
	defer lock.Release(ctx)
	fmt.Println("I have a lock!")

	// Sleep and check the remaining TTL.
	time.Sleep(50 * time.Millisecond)
	if ttl, err := lock.TTL(ctx); err != nil {
		log.Fatalln(err)
	} else if ttl > 0 {
		fmt.Println("Yay, I still have my lock!")
	}

	lock2, err := locker.Obtain(ctx, "my-key", 150*time.Millisecond, opts)
	if err == redislock.ErrNotObtained {
		fmt.Println("Could not obtain lock!")
	} else if err != nil {
		log.Fatalln(err)
	}

	// Sleep and check the remaining TTL.
	time.Sleep(50 * time.Millisecond)
	if ttl, err := lock2.TTL(ctx); err != nil {
		log.Fatalln(err)
	} else if ttl > 0 {
		fmt.Println("Yay, I still have my lock!")
	}

	lock3, err := locker.Obtain(ctx, "my-key", 100*time.Millisecond, nil)
	if err == redislock.ErrNotObtained {
		fmt.Println("Could not obtain lock, but it's ok!")
	} else if err != nil {
		log.Fatalln(err)
	}

	if lock3 != nil {
		log.Fatalln("lock3 should be nil")
	}

	// Sleep a little longer, then check.
	time.Sleep(100 * time.Millisecond)
	if ttl, err := lock.TTL(ctx); err != nil {
		log.Fatalln(err)
	} else if ttl == 0 {
		fmt.Println("Now, my lock has expired!")
	}

	// Output:
	// I have a lock!
	// Yay, I still have my lock!
	// Yay, I still have my lock!
	// Could not obtain lock, but it's ok!
	// Now, my lock has expired!
}

func ExampleClient_Obtain_retry() {
	client := redis.NewClient(&redis.Options{Network: "tcp", Addr: "127.0.0.1:6379"})
	defer client.Close()

	locker := redislock.New(client)

	ctx := context.Background()

	// Retry every 100ms, for up-to 3x
	backoff := redislock.LimitRetry(redislock.LinearBackoff(100*time.Millisecond), 3)

	// Obtain lock with retry
	lock, err := locker.Obtain(ctx, "my-key", time.Second, &redislock.Options{
		RetryStrategy: backoff,
	})
	if err == redislock.ErrNotObtained {
		fmt.Println("Could not obtain lock!")
	} else if err != nil {
		log.Fatalln(err)
	}
	defer lock.Release(ctx)

	fmt.Println("I have a lock!")
}

func ExampleClient_Obtain_customDeadline() {
	client := redis.NewClient(&redis.Options{Network: "tcp", Addr: "127.0.0.1:6379"})
	defer client.Close()

	locker := redislock.New(client)

	// Retry every 500ms, for up-to a minute
	backoff := redislock.LinearBackoff(500 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	defer cancel()

	// Obtain lock with retry + custom deadline
	lock, err := locker.Obtain(ctx, "my-key", time.Second, &redislock.Options{
		RetryStrategy: backoff,
	})
	if err == redislock.ErrNotObtained {
		fmt.Println("Could not obtain lock!")
	} else if err != nil {
		log.Fatalln(err)
	}
	defer lock.Release(context.Background())

	fmt.Println("I have a lock!")
}
