package main

import (
	"sync"
)

type Call struct {
	err    error
	data   any
	waiter sync.WaitGroup
}

type board struct {
	argsGuard sync.Mutex
	args      map[string]*Call
}

func NewBoard() *board {
	return &board{
		args: make(map[string]*Call),
	}
}

// singleflight concurrency pattern
func (b *board) GetUser(user string) any {
	b.argsGuard.Lock()
	if user, ok := b.args[user]; ok {
		b.argsGuard.Unlock()
		user.waiter.Wait()
		return user.data
	}

	c := new(Call)
	c.waiter.Add(1)
	b.args[user] = c
	b.argsGuard.Unlock()

	c.data = expensiveCalc()
	c.err = nil
	c.waiter.Done()

	b.argsGuard.Lock()
	delete(b.args, user)
	b.argsGuard.Unlock()

	return c.data
}

func expensiveCalc() string {
	// some time-consuming calculation
	return "some expensive data"
}
