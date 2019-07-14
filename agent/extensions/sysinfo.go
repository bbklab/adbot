package extensions

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cloudfoundry/gosigar"
	"github.com/docker/docker/pkg/parsers/kernel"
	"github.com/docker/docker/pkg/parsers/operatingsystem"

	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/types"
)

var (
	dockerSocket = "/var/run/docker.sock"
	dockerClient *http.Client // note: reuse the client to avoid leaks
)

func init() {
	if env := os.Getenv("DOCKER_SOCKET_PATH"); env != "" {
		dockerSocket = env
	}

	dockerClient = &http.Client{
		Transport: &http.Transport{
			Dial: func(_, _ string) (net.Conn, error) {
				return net.DialTimeout("unix", dockerSocket, time.Second*5)
			},
		},
	}
}

// GatherSysInfo is exported
func GatherSysInfo() (*types.SysInfo, error) {
	g := new(gatherer)
	funcs := []gatherFunc{
		g.operatingSystem,
		g.kernel,
		g.hostName,
		g.uptime,
		g.unixTime,
		g.loadAvgs,
		g.cpu,
		g.memory,
		g.swap,
		g.user,
		g.ips,
		g.disks,
		g.diskstats,
		g.traffics,
		g.docker,
		g.bbr,
		g.withMaster,
		g.manufacturer,
	}

	runfWithTimeout := func(f func() error) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		ch := make(chan error, 1)
		go func() {
			ch <- f()
		}()

		select {
		case err := <-ch:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	for _, fun := range funcs {
		if err := runfWithTimeout(fun); err != nil {
			log.Errorf("GatherSysInfo.%v error: %v", utils.FuncName(fun), err) // log error and continue ~
		}
	}

	return &g.info, nil
}

type gatherFunc func() error

type gatherer struct {
	info types.SysInfo
}

func (g *gatherer) operatingSystem() error {
	operatingSystem := "<unknown>"
	if s, err := operatingsystem.GetOperatingSystem(); err == nil {
		operatingSystem = s
	}
	g.info.OS = operatingSystem
	return nil
}

func (g *gatherer) kernel() error {
	kernelVersion := "<unknown>"
	if kv, err := kernel.GetKernelVersion(); err == nil {
		kernelVersion = kv.String()
	}
	g.info.Kernel = kernelVersion
	return nil
}

func (g *gatherer) hostName() error {
	name, err := os.Hostname()
	if err != nil {
		return err
	}

	g.info.Hostname = name
	return nil
}

func (g *gatherer) uptime() error {
	up := sigar.Uptime{}
	if err := up.Get(); err != nil {
		return err
	}

	g.info.Uptime = strconv.FormatFloat(up.Length, 'f', 6, 64)
	g.info.UptimeInt = int64(up.Length)
	return nil
}

func (g *gatherer) unixTime() error {
	g.info.UnixTime = time.Now().Unix()
	return nil
}

func (g *gatherer) loadAvgs() error {
	avg := sigar.LoadAverage{}
	if err := avg.Get(); err != nil {
		return err
	}

	g.info.LoadAvgs = types.LoadAvgInfo{
		One:     avg.One,
		Five:    avg.Five,
		Fifteen: avg.Fifteen,
	}
	return nil
}

func (g *gatherer) cpu() error {
	// cpu used %
	//  - collect two samples
	//  - get the delta between two samples
	//  - caculate the cpu used %
	var used uint64
	samples := [2]sigar.Cpu{}
	for i := 0; i <= 1; i++ {
		err := samples[i].Get()
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 500)
	}
	deltaCPU := samples[1].Delta(samples[0])
	used = 100 - (deltaCPU.Idle*100)/deltaCPU.Total()

	// nb of cpu processor
	var processorN int64
	cpulist := sigar.CpuList{}
	if err := cpulist.Get(); err != nil {
		return err
	}
	processorN = int64(len(cpulist.List))

	// nb of physical cpu
	var physicalN int64
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return err
	}
	defer f.Close()

	var (
		scanner = bufio.NewScanner(f)
		reg     = regexp.MustCompile(`^physical id\s+:\s+(\d+)$`)
	)
	for scanner.Scan() {
		subMatch := reg.FindSubmatch([]byte(scanner.Text()))
		if len(subMatch) >= 2 {
			physicalN++
		}
	}

	// final
	g.info.CPU = types.CPUInfo{
		Processor: processorN,
		Physical:  physicalN,
		Used:      used,
	}
	return nil
}

func (g *gatherer) memory() error {
	mem := sigar.Mem{}
	if err := mem.Get(); err != nil {
		return err
	}

	g.info.Memory = types.MemoryInfo{
		Total:  int64(mem.Total),
		Used:   int64(mem.ActualUsed),
		Cached: int64(mem.Used - mem.ActualUsed),
	}
	return nil
}

func (g *gatherer) swap() error {
	swap := sigar.Swap{}
	if err := swap.Get(); err != nil {
		return err
	}

	g.info.Swap = types.SwapInfo{
		Total: int64(swap.Total),
		Used:  int64(swap.Used),
		Free:  int64(swap.Free),
	}
	return nil
}

func (g *gatherer) user() error {
	// in docker container, sometimes can't obtain UserName by os.Getenv("USER"),
	// which lead user.Current() complains: Current not implemented on linux/amd64
	current := types.UserInfo{
		UID: strconv.Itoa(os.Getuid()),
		GID: strconv.Itoa(os.Getgid()),
	}

	user, err := user.LookupId(current.UID)
	if err != nil {
		return err
	}
	current.Name = user.Username

	if current.UID != "0" {
		current.Sudo = checkSudo()
	}

	g.info.User = current
	return nil
}

// utils
//
func processFile(file string, handler func(string)) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				if len(line) > 0 {
					goto PROCESS
				}
				break
			}
			return err
		}
	PROCESS:
		line = bytes.TrimSuffix(line, []byte("\n")) // trim final '\n'
		if bytes.HasSuffix(line, []byte("\r")) {    // trim possible final '\r'
			line = bytes.TrimSuffix(line, []byte("\r"))
		}
		handler(string(line))
	}

	return nil
}

func checkSudo() bool {
	cmd := exec.Command("sudo", "-n", "echo", "sudo test")
	_, err := cmd.CombinedOutput()
	return err == nil // we're expecting `user ALL=(ALL) NOPASSWD: ALL`, otherwise error message like: sudo: a password is required
}
