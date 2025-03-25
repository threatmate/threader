package threader

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
)

// Threader is a wrapper around go routines that builds in wait groups and handles panics.
// You can wait for all of the threads in a Threader to complete if you want, and when done,
// you can ask if any of them failed.
//
// Typically, multiple threads will attempt to modify some shared state.  You may use the
// Lock and Unlock functions to guard this state.
type Threader struct {
	wg          sync.WaitGroup // This is the wait group that waits for all threads to finish.
	mutex       sync.Mutex     // This is the mutex that guards the errs slice.
	errs        []error        // This is a list of errors from the running threads.
	publicMutex sync.Mutex     // This is the public mutex that can be used to guard shared state; the caller can use this instead of creating their own mutex.  The Threader itself will never use this.
}

// DefaultThreader is the default Threader.
//
// This can be used to defend against panics without having to manage a Threader.
// However, it is best practice to create a Threader instance.
var DefaultThreader = &Threader{}

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
func (r *Threader) Go(fn func()) {
	r.GoWithErr(func() error { fn(); return nil })
}

// GoWithErr runs a function in a goroutine.
func (r *Threader) GoWithErr(fn func() error) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		defer func() {
			// recover from panic if one occured. Set err to nil otherwise.
			// If we don't recover from a panic in a goroutine the entire app will crash.
			// This also includes the main thread.
			if value := recover(); value != nil {
				r.mutex.Lock()
				r.errs = append(r.errs, fmt.Errorf("panic: %v; stacktrace: %s", value, string(debug.Stack())))
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
