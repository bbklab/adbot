package main

import (
	"fmt"
	"io"
	"os"

	log "github.com/Sirupsen/logrus"
	goadb "github.com/zach-klippenstein/goadb"
)

func main() {
	adb, err := goadb.New()
	if err != nil {
		log.Fatalln(err)
	}

	adb.StartServer()

	version, err := adb.ServerVersion()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(version)

	// adb devices -l
	devices, err := adb.ListDevices()
	if err != nil {
		log.Fatalln(err)
	}

	for _, d := range devices {
		log.Println(d.Serial, d.Product, d.Model, d.Usb)

		var (
			dvc = adb.Device(goadb.DeviceWithSerial(d.Serial))
		)

		// dev serial
		serialNo, err := dvc.Serial()
		if err != nil {
			log.Fatalln(err)
		}

		// dev path
		devPath, err := dvc.DevicePath()
		if err != nil {
			log.Fatalln(err)
		}

		// dev state
		state, err := dvc.State()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(dvc)
		fmt.Printf("\tserial no: %s\n", serialNo)
		fmt.Printf("\tdevPath: %s\n", devPath)
		fmt.Printf("\tstate: %s(%t)\n", state, state == goadb.StateOnline)

		// stat file
		stat, err := dvc.Stat("/sdcard")
		if err != nil {
			fmt.Println("\terror stating /sdcard:", err)
		}
		fmt.Printf("\tstat \"/sdcard\": %+v\n", stat)

		stat, err = dvc.Stat("/notexists.files.xxxx")
		if err != nil {
			fmt.Println("\terror stating /notexists.files.xxxx:", err)
		} else {
			fmt.Printf("\tstat \"/notexists.files.xxxx\": %+v\n", stat)
		}

		// read file
		fmt.Print("\tload avg: ")
		loadavgReader, err := dvc.OpenRead("/proc/loadavg")
		if err != nil {
			log.Fatalln(err)
		}
		defer loadavgReader.Close()
		io.Copy(os.Stdout, loadavgReader)

		// run command
		bs, err := dvc.RunCommand("input", "keyevent", "26")
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(string(bs))

		bs, err = dvc.RunCommand("dumpsys", "window", "policy", "|", "grep", "-E", "mScreenOnEarly=true")
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(string(bs))
	}
}
