package queue

import "errors"

// Put adds a slice of bytes to a channel queue. If the channel isn't buffered
// this will block.
func Put(queue chan<- byte, items []byte) {
	for _, i := range items {
		queue <- i
	}
}

// Get removes an exact number of items from the channel and blocks until that
// number of items are available. If the channel is closed then an error is
// returned.
func Get(queue <-chan byte, number int) ([]byte, error) {
	var result []byte

ReadLoop:
	for {
		select {
		case i, ok := <-queue:
			if ok {
				result = append(result, i)
				if len(result) == number {
					break ReadLoop
				}
			} else {
				return nil, errors.New("queue channel is closed")
			}
		}
	}

	return result, nil
}
