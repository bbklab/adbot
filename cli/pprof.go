package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/urfave/cli"

	"github.com/bbklab/adbot/cli/helpers"
	"github.com/bbklab/adbot/pkg/color"
)

var (
	dumpPProfFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "output,o",
			Usage: "FILE  Write output to <file> instead of stdout",
		},
		cli.IntFlag{
			Name:  "seconds,s",
			Usage: "duration for time-based profile (eg: cpu,trace) collection, by seconds",
			Value: 30,
		},
	}
)

// PProfCommand is exported
func PProfCommand() cli.Command {
	return cli.Command{
		Name:  "pprof",
		Usage: "dump pprof datas",
		Subcommands: []cli.Command{
			cpuCommand(),          // cpu
			traceCommand(),        // trace
			goroutineCommand(),    // goroutine
			mutexCommand(),        // mutex
			blockCommand(),        // block
			threadcreateCommand(), // threadcreate
			heapCommand(),         // heap
		},
	}
}

func cpuCommand() cli.Command {
	return cli.Command{
		Name:   "cpu",
		Usage:  "dump cpu profiling datas",
		Flags:  dumpPProfFlags,
		Action: saveCPUPprof,
	}
}

func traceCommand() cli.Command {
	return cli.Command{
		Name:   "trace",
		Usage:  "dump trace profiling datas",
		Flags:  dumpPProfFlags,
		Action: saveTracePprof,
	}
}

func goroutineCommand() cli.Command {
	return cli.Command{
		Name:   "goroutine",
		Usage:  "dump goroutine profiling datas",
		Flags:  dumpPProfFlags,
		Action: saveGoroutinePprof,
	}
}

func mutexCommand() cli.Command {
	return cli.Command{
		Name:   "mutex",
		Usage:  "dump mutex profiling datas",
		Flags:  dumpPProfFlags,
		Action: saveMutexPprof,
	}
}

func blockCommand() cli.Command {
	return cli.Command{
		Name:   "block",
		Usage:  "dump block profiling datas",
		Flags:  dumpPProfFlags,
		Action: saveBlockPprof,
	}
}

func threadcreateCommand() cli.Command {
	return cli.Command{
		Name:   "threadcreate",
		Usage:  "dump threadcreate profiling datas",
		Flags:  dumpPProfFlags,
		Action: saveThreadCreatePprof,
	}
}

func heapCommand() cli.Command {
	return cli.Command{
		Name:   "heap",
		Usage:  "dump heap profiling datas",
		Flags:  dumpPProfFlags,
		Action: saveHeapPprof,
	}
}

func saveCPUPprof(c *cli.Context) error {
	os.Stdout.Write(append([]byte(color.Cyan("NOTICE: better to give some pressure against the server at the same time")), '\r', '\n'))
	return doSavePprofData("profile", c)
}
func saveTracePprof(c *cli.Context) error        { return doSavePprofData("trace", c) }
func saveGoroutinePprof(c *cli.Context) error    { return doSavePprofData("goroutine", c) }
func saveMutexPprof(c *cli.Context) error        { return doSavePprofData("mutex", c) }
func saveBlockPprof(c *cli.Context) error        { return doSavePprofData("block", c) }
func saveThreadCreatePprof(c *cli.Context) error { return doSavePprofData("threadcreate", c) }
func saveHeapPprof(c *cli.Context) error         { return doSavePprofData("heap", c) }

func doSavePprofData(pname string, c *cli.Context) error {
	client, err := helpers.NewClient()
	if err != nil {
		return err
	}

	var (
		foutput = c.String("output")
		seconds = c.Int("seconds")
	)

	if foutput == "" {
		foutput = fmt.Sprintf("pprof.%s.samples.%s.%d.pb.gz", "adbot", pname, time.Now().Unix()) // default output file name
	}

	stream, err := client.PProfData(pname, seconds)
	if err != nil {
		return err
	}
	defer stream.Close()

	fd, err := os.OpenFile(foutput, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()

	n, err := io.Copy(fd, stream)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, fmt.Sprintf("%d bytes, %s", n, foutput))
	return nil
}
