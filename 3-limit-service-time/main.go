//////////////////////////////////////////////////////////////////////
//
// Your video processing service has a freemium model. Everyone has 10
// sec of free processing time on your service. After that, the
// service will kill your process, unless you are a paid premium user.
//
// Beginner Level: 10s max per request
// Advanced Level: 10s max per user (accumulated)
//

package main

import (
	"context"
	"fmt"
	"time"
)

// User defines the UserModel. Use this to check whether a User is a
// Premium user or not
type User struct {
	ID        int
	IsPremium bool
	TimeUsed  int64 // in seconds
}

// HandleRequest runs the processes requested by users. Returns false
// if process had to be killed
func HandleRequest(process func(), u *User) bool {
	if u.IsPremium {
		// Premium users have no restrictions
		process()
		return true
	}

	remainingTime := 10
	if remainingTime <= 0 {
		// User has no remaining time
		return false
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(remainingTime)*time.Second)
	defer cancelFunc()

	done := make(chan bool)

	startTime := time.Now()
	go func() {
		process()
		done <- true
	}()

	defer fmt.Printf("User %v", u)
	select {
	case <-ctx.Done():
		u.TimeUsed += int64(time.Since(startTime).Seconds())
		return false
	case <-done:
		u.TimeUsed += int64(time.Since(startTime).Seconds())
		return true
	}
}

func main() {
	RunMockServer()
}
