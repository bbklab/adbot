package debug

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/pkg/template"
	"github.com/bbklab/adbot/pkg/utils"
)

func init() {
	debugC := make(chan os.Signal, 1)
	signal.Notify(debugC, syscall.SIGUSR1)
	ftrace := filepath.Join(os.TempDir(), "adbot-stack-trace.log")

	go func() {
		for range debugC {
			f, err := os.OpenFile(ftrace, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				logrus.Error("write stack trace log file error", err)
				break
			}

			fmt.Fprint(f, "GENERAL\n\n")
			NewDebugInfo().WriteTo(f)

			fmt.Fprint(f, "\n\nGOROUTINE\n\n")
			pprof.Lookup("goroutine").WriteTo(f, 2)

			fmt.Fprint(f, "\n\nHEAP\n\n")
			pprof.Lookup("heap").WriteTo(f, 1)

			fmt.Fprint(f, "\n\nTHREADCREATE\n\n")
			pprof.Lookup("threadcreate").WriteTo(f, 1)

			fmt.Fprint(f, "\n\nBLOCK\n\n")
			pprof.Lookup("block").WriteTo(f, 1)

			f.Close()
		}
	}()
}

// MemStats dump the runtime memory stats
func MemStats() runtime.MemStats {
	s := new(runtime.MemStats)
	runtime.ReadMemStats(s)
	return *s
}

// Info is exported
type Info struct {
	UnixTime      int64      `json:"time"`
	Os            string     `json:"os"`
	Arch          string     `json:"arch"`
	GoVersion     string     `json:"go_version"`
	MaxProcs      int64      `json:"max_procs"`
	NumCpus       int64      `json:"num_cpus"`
	NumGoroutines int64      `json:"num_goroutines"`
	NumCgoCalls   int64      `json:"num_cgocalls"`
	NumFds        int64      `json:"num_fds"`
	Memory        MemoryInfo `json:"memory"`
}

// MemoryInfo is exported
type MemoryInfo struct {
	// memory
	MemoryAlloc      uint64 `json:"memory_alloc"`       // bytes
	MemoryTotalAlloc uint64 `json:"memory_total_alloc"` // bytes
	MemorySys        uint64 `json:"memory_sys"`         // bytes
	MemoryLookups    uint64 `json:"memory_lookups"`     // nb
	MemoryMallocs    uint64 `json:"memory_mallocs"`     // nb
	MemoryFrees      uint64 `json:"memory_frees"`       // nb

	// stack
	StackInUse uint64 `json:"stack_inuse"` // bytes

	// heap
	HeapAlloc    uint64 `json:"heap_alloc"`    // bytes
	HeapSys      uint64 `json:"heap_sys"`      // bytes
	HeapIdle     uint64 `json:"heap_idle"`     // bytes
	HeapInuse    uint64 `json:"heap_inuse"`    // bytes
	HeapReleased uint64 `json:"heap_released"` // bytes
	HeapObjects  uint64 `json:"heap_objects"`  // nb
}

// NewDebugInfo is exported
func NewDebugInfo() *Info {
	mem := MemStats()

	return &Info{
		UnixTime:      time.Now().Unix(),
		Os:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		GoVersion:     runtime.Version(),
		MaxProcs:      int64(runtime.GOMAXPROCS(0)),
		NumCpus:       int64(runtime.NumCPU()),
		NumGoroutines: int64(runtime.NumGoroutine()),
		NumCgoCalls:   int64(runtime.NumCgoCall()),
		NumFds:        int64(utils.NumFd()),
		Memory: MemoryInfo{
			MemoryAlloc:      mem.Alloc,
			MemoryTotalAlloc: mem.TotalAlloc,
			MemorySys:        mem.Sys,
			MemoryLookups:    mem.Lookups,
			MemoryMallocs:    mem.Mallocs,
			MemoryFrees:      mem.Frees,
			StackInUse:       mem.StackInuse,
			HeapAlloc:        mem.HeapAlloc,
			HeapSys:          mem.HeapSys,
			HeapIdle:         mem.HeapIdle,
			HeapInuse:        mem.HeapInuse,
			HeapReleased:     mem.HeapReleased,
			HeapObjects:      mem.HeapObjects,
		},
	}
}

var debugInfoTemplate = ` UnixTime:      {{.UnixTime}}
 OS/Arch:       {{.Os}}/{{.Arch}}
 Go version:    {{.GoVersion}}
 MaxProcs:      {{.MaxProcs}}
 NumCpus:       {{.NumCpus}}
 NumGoroutines: {{.NumGoroutines}}
 NumCgoCalls:   {{.NumCgoCalls}}
 NumFds:        {{.NumFds}}
 Memory:
   MemoryAlloc:      {{size .Memory.MemoryAlloc}}
   MemoryTotalAlloc: {{size .Memory.MemoryTotalAlloc}}
   MemorySys:        {{size .Memory.MemorySys}}
   MemoryLookups:    {{.Memory.MemoryLookups}}
   MemoryMallocs:    {{.Memory.MemoryMallocs}}
   MemoryFrees:      {{.Memory.MemoryFrees}}
   StackInUse:       {{size .Memory.StackInUse}}
   HeapAlloc:        {{size .Memory.HeapAlloc}}
   HeapSys:          {{size .Memory.HeapSys}}
   HeapIdle:         {{size .Memory.HeapIdle}}
   HeapInuse:        {{size .Memory.HeapInuse}}
   HeapReleased:     {{size .Memory.HeapReleased}}
   HeapObjects:      {{.Memory.HeapObjects}}
`

// WriteTo is exported
func (info *Info) WriteTo(w io.Writer) (int64, error) {
	tmpl, err := template.NewParser(debugInfoTemplate)
	if err != nil {
		return -1, err
	}
	return -1, tmpl.Execute(w, info) // just make pass govet
}
