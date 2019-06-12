package main

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/bbklab/paybot/pkg/axel"
)

func main() {
	// dl, err := axel.New("http://a.sina.com.cn:82/tmp/100M.k", "100M", 10, time.Second*10)
	// dl, err := axel.New("http://bbklab.net:81/tmp/1B", "1B.file", 10, time.Second*10)
	dl, err := axel.New("http://bbklab.net:81/tmp/100M", "100M.file", 10, time.Second*10)
	if err != nil {
		logrus.Fatalln(err)
	}

	err = dl.Download()
	if err != nil {
		logrus.Fatalln(err)
	}

	fmt.Println("OK")
}
