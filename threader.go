package threader

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Threader is a wrapper around go routines that builds in wait groups and handles panics.
// You can wait for all of the threads in a Threader to complete if you want, and when done,
// you can ask if any of them failed.
//
// Typically, multiple threads will attempt to modify some shared state.  You may use the
// Lock and Unlock functions to guard this state.
type Threader struct {
	wg          sync.WaitGroup
	mutex       sync.Mutex
	errs        []error
	publicMutex sync.Mutex
}

// DefaultThreader is the default Threader.
var DefaultThreader = &Threader{}

// Go runs a go routine using the default Threader.
func Go(ctx context.Context, fn func()) {
	DefaultThreader.Go(ctx, fn)
}

// GoWithErr runs a go routine using the default Threader.
func GoWithErr(ctx context.Context, fn func() error) {
	DefaultThreader.GoWithErr(ctx, fn)
}

// New returns a new Threader.
func New() *Threader {
	return &Threader{}
}

// Lock the public mutex.
//
// This lock may be used to guard shared state.
func (r *Threader) Lock() {
	r.publicMutex.Lock()
}

// Unlock the public mutex.
//
// This lock may be used to guard shared state.
func (r *Threader) Unlock() {
	r.publicMutex.Unlock()
}

// Go run a function in a goroutine.
func (r *Threader) Go(ctx context.Context, fn func()) {
	r.GoWithErr(ctx, func() error { fn(); return nil })
}

// GoWithErr runs a function in a goroutine.
func (r *Threader) GoWithErr(ctx context.Context, fn func() error) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		defer func() {
			// recover from panic if one occured. Set err to nil otherwise.
			// If we don't recover from a panic in a goroutine the entire app will crash.
			// This also includes the main thread.
			if value := recover(); value != nil {
				r.mutex.Lock()
				r.errs = append(r.errs, fmt.Errorf("panic: %v", value))
				r.mutex.Unlock()
			}
		}()

		err := fn()
		if err != nil {
			r.mutex.Lock()
			r.errs = append(r.errs, err)
			r.mutex.Unlock()
		}
	}()
}

// Wait waits for all go routines to complete and then returns an error if at least one
// go routine had an error.
//
// After calling this function, all errors will be cleared.
func (r *Threader) Wait() error {
	r.wg.Wait() // Wait for everyone to finish.

	r.mutex.Lock()
	defer r.mutex.Unlock()

	var output error
	if len(r.errs) > 0 {
		output = errors.Join(r.errs...)
	}

	r.errs = []error{} // Reset the errors, since we're done running.

	return output
}
