package di

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"
)

type closer struct {
	noErrCloser func()
}

func (c *closer) Close() error {
	c.noErrCloser()

	return nil
}

func NewCloser(noErrCloser func()) io.Closer {
	return &closer{noErrCloser}
}

type closerFunc struct {
	key    string
	closer io.Closer
}

var closerFuncs = make([]*closerFunc, 0)

func CloseAll() {
	if len(closerFuncs) == 0 {
		log.Println("no closer registered")

		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		for i := len(closerFuncs) - 1; i >= 0; i-- {
			c := closerFuncs[i]
			if err := c.closer.Close(); err != nil {
				log.Println(fmt.Sprintf("failed to close %s: %v", c.key, err))
			}

			log.Println(fmt.Sprintf("successfully closed %s", c.key))
		}

		done <- nil
	}()

	select {
	case <-ctx.Done():
		log.Println("timeout before finishing all clean up processes")
	case <-done:
		return
	}
}

func RegisterCloser(key string, closer io.Closer) {
	for _, c := range closerFuncs {
		if key == c.key {
			log.Fatal(fmt.Sprintf("duplicate closer key: %s", key))
		}
	}

	closerFuncs = append(closerFuncs, &closerFunc{
		key:    key,
		closer: closer,
	})

	log.Println(fmt.Sprintf("%s closer is registered", key))
}
