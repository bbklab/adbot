package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/bbklab/paybot/pkg/rate"
)

func main() {
	// init limit 1/s
	l := rate.NewLimiter(time.Second*1, 1)
	fmt.Println(l)

	// after 3s, we allow 2/s
	time.AfterFunc(time.Second*3, func() {
		l.SetLimit(time.Second*1, 2)
	})

	// then we launch 10 tasks
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 1; i <= 10; i++ {
		go func(i int) {
			defer wg.Done()

			// try to take one token
			for {
				err := l.Take()
				if err == nil {
					break
				}
				time.Sleep(time.Millisecond * 100)
			}

			logrus.Println("------->", i, "RUNNING")
		}(i)
	}
	wg.Wait()
	logrus.Println("STEP1 ALL DONE")
}
