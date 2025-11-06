package jobs

import "sync"

type MultiTasker struct {
	wg     sync.WaitGroup
	errors chan error
}

func NewMultiTasker() *MultiTasker {
	return &MultiTasker{
		errors: make(chan error),
	}
}

func (mt *MultiTasker) Queue(f func() error) {
	mt.wg.Add(1)
	go func() {
		defer mt.wg.Done()
		mt.errors <- f()
	}()
}

func (mt *MultiTasker) Wait(errCallback func(error)) error {
	go func() {
		defer close(mt.errors)
		mt.wg.Wait()
	}()
	var rvErr error
	for err := range mt.errors {
		if err != nil && rvErr == nil {
			rvErr = err
			errCallback(err)
		}
	}
	return rvErr
}
