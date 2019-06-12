package scheduler

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
)

var (
	namespace = "adbot"
)

func init() {
	prometheus.MustRegister(newGlobalExporter())
}

type globalExporter struct {
	scrapeCost  prometheus.Gauge // scrape count & cost
	scrapeCount prometheus.Counter
	exporters   map[string]prometheus.Collector // various subsystem collectors (exporters)
}

func newGlobalExporter() *globalExporter {
	e := &globalExporter{
		scrapeCost: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "scrape_cost_seconds",
			Help:      "current scrape duration cost by seconds.",
		}),
		scrapeCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "scrape_count",
			Help:      "total count of scrape.",
		}),
	}

	e.exporters = map[string]prometheus.Collector{
		"node": newNodeExporter(), // node exporter
	}

	return e
}

type nodeExporter struct {
	count           prometheus.Gauge // count & status
	status          *prometheus.GaugeVec
	latency         *prometheus.GaugeVec
	loadAvg         *prometheus.GaugeVec // loadavg
	cpuUsage        *prometheus.GaugeVec // cpu
	memoryTotal     *prometheus.GaugeVec // memory
	memoryUsed      *prometheus.GaugeVec
	memoryCached    *prometheus.GaugeVec
	swapTotal       *prometheus.GaugeVec // swap
	swapUsed        *prometheus.GaugeVec
	swapFree        *prometheus.GaugeVec
	diskTotal       *prometheus.GaugeVec // disk space
	diskUsed        *prometheus.GaugeVec
	diskFree        *prometheus.GaugeVec
	diskINode       *prometheus.GaugeVec
	diskIFree       *prometheus.GaugeVec
	networkRx       *prometheus.GaugeVec // network traffic
	networkTx       *prometheus.GaugeVec
	diskIORead      *prometheus.GaugeVec // disk io
	diskIOWrite     *prometheus.GaugeVec
	containersCount *prometheus.GaugeVec // containers
}

func newNodeExporter() *nodeExporter {
	var (
		subsystem = "node"
	)

	return &nodeExporter{
		// count & status
		count: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "count_total",
			Help:      "total count of current adbot nodes.",
		}),
		status: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "status",
			Help:      "status of current adbot nodes, 0:normal 1:abnormal",
		}, []string{"node_id", "hostname", "remote", "cpu", "memory"}),

		// node latency
		latency: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "latency_seconds",
			Help:      "networking latency between node & master communication, by seconds",
		}, []string{"node_id", "hostname", "remote"}),

		// node loadavg
		loadAvg: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "load_avgerage",
			Help:      "node load average in one minute",
		}, []string{"node_id", "hostname", "remote"}),

		// node cpu usage
		cpuUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cpu_usage",
			Help:      "node current cpu usage, by %",
		}, []string{"node_id", "hostname", "remote"}),

		// node memory
		memoryTotal: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "memory_bytes_total",
			Help:      "node total memory in bytes",
		}, []string{"node_id", "hostname", "remote"}),
		memoryUsed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "memory_bytes_used",
			Help:      "node used memory in bytes",
		}, []string{"node_id", "hostname", "remote"}),
		memoryCached: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "memory_bytes_cached",
			Help:      "node cached memory in bytes",
		}, []string{"node_id", "hostname", "remote"}),

		// node swap
		swapTotal: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "swap_bytes_total",
			Help:      "node total swap in bytes",
		}, []string{"node_id", "hostname", "remote"}),
		swapUsed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "swap_bytes_used",
			Help:      "node used swap in bytes",
		}, []string{"node_id", "hostname", "remote"}),
		swapFree: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "swap_bytes_free",
			Help:      "node free swap in bytes",
		}, []string{"node_id", "hostname", "remote"}),

		// node disk
		diskTotal: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "disk_bytes_total",
			Help:      "node disk total in bytes",
		}, []string{"node_id", "hostname", "remote", "dev_name", "mount_at"}),
		diskUsed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "disk_bytes_used",
			Help:      "node disk used in bytes",
		}, []string{"node_id", "hostname", "remote", "dev_name", "mount_at"}),
		diskFree: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "disk_bytes_free",
			Help:      "node disk free in bytes",
		}, []string{"node_id", "hostname", "remote", "dev_name", "mount_at"}),
		diskINode: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "disk_inode_total",
			Help:      "node disk inode total count",
		}, []string{"node_id", "hostname", "remote", "dev_name", "mount_at"}),
		diskIFree: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "disk_inode_free",
			Help:      "node disk inode free count",
		}, []string{"node_id", "hostname", "remote", "dev_name", "mount_at"}),

		// node network traffic
		networkRx: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "traffic_bytes_rx",
			Help:      "node network traffic rx in bytes",
		}, []string{"node_id", "hostname", "remote", "device"}),
		networkTx: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "traffic_bytes_tx",
			Help:      "node network traffic tx in bytes",
		}, []string{"node_id", "hostname", "remote", "device"}),

		// node diskio stats
		diskIORead: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "diskio_bytes_read",
			Help:      "node diskio read in bytes",
		}, []string{"node_id", "hostname", "remote", "device"}),
		diskIOWrite: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "diskio_bytes_write",
			Help:      "node diskio write in bytes",
		}, []string{"node_id", "hostname", "remote", "device"}),

		// node container count
		containersCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "containers_count",
			Help:      "total count of current node running containers",
		}, []string{"node_id", "hostname", "remote"}),
	}
}

//
// globalExporter implemention
//

func (m *globalExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.scrapeCost.Desc()
	ch <- m.scrapeCount.Desc()
	for _, e := range m.exporters {
		e.Describe(ch)
	}
}

func (m *globalExporter) Collect(ch chan<- prometheus.Metric) {
	var startAt = time.Now()

	defer func() {
		m.scrapeCost.Set(time.Since(startAt).Seconds())
		ch <- m.scrapeCost

		m.scrapeCount.Inc()
		ch <- m.scrapeCount
	}()

	var wg sync.WaitGroup
	wg.Add(len(m.exporters))
	for _, e := range m.exporters {
		go func(e prometheus.Collector) {
			e.Collect(ch)
			wg.Done()
		}(e)
	}
	wg.Wait()
}

//
// nodeExporter implemention
//

func (m *nodeExporter) Describe(ch chan<- *prometheus.Desc) {
}

// TODO: use cache.go to export metrics
// this maybe very expensive while enormous count of nodes
func (m *nodeExporter) Collect(ch chan<- prometheus.Metric) {
	// get db nodes list
	nodes, err := store.DB().ListNodes(nil)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(
			prometheus.NewDesc("collect_db_nodes_failed", "can't collect db nodes", nil, nil),
			fmt.Errorf("list db nodes error: %v", err),
		)
		return
	}
	m.count.Set(float64(len(nodes)))
	ch <- m.count

	// concurrency send the node metrics
	var (
		wg     sync.WaitGroup
		tokens = make(chan struct{}, runtime.NumCPU()*2) // nb of max concurrency
	)

	wg.Add(len(nodes))
	for _, node := range nodes {
		tokens <- struct{}{} // take a token to start up a new worker, haninig block if tokens channel is full

		go func(node *types.Node) {

			defer func() {
				<-tokens // release one token
				wg.Done()
			}()

			// node status
			lbs := prometheus.Labels{ // must be full matched as defined to avoid panic: Inconsistent label cardinality
				"node_id":  node.ID,
				"remote":   node.RemoteIP(),
				"hostname": "-",
				"cpu":      "0",
				"memory":   "0",
			}
			if info := node.SysInfo; info != nil {
				lbs["hostname"] = info.Hostname
				lbs["cpu"] = fmt.Sprintf("%d", info.CPU.Processor)
				lbs["memory"] = fmt.Sprintf("%d", info.Memory.Total)
			}
			metric := m.status.With(lbs)

			if node.Status == types.NodeStatusOnline {
				metric.Set(float64(0))
			} else {
				metric.Set(float64(1))
			}

			ch <- metric

			// note: if node not online, do NOT export this node's metric
			if node.Status != types.NodeStatusOnline {
				return
			}

			// note: if node hasn't sysinfo ready, do NOT export this node's metric
			info := node.SysInfo
			if info == nil {
				return
			}

			lbs = prometheus.Labels{
				"node_id":  node.ID,
				"remote":   node.RemoteIP(),
				"hostname": info.Hostname,
			}

			metric = m.latency.With(lbs)
			metric.Set(float64(node.Latency.Seconds()))
			ch <- metric

			metric = m.loadAvg.With(lbs)
			metric.Set(float64(info.LoadAvgs.One))
			ch <- metric

			metric = m.cpuUsage.With(lbs)
			metric.Set(float64(info.CPU.Used))
			ch <- metric

			metric = m.memoryTotal.With(lbs)
			metric.Set(float64(info.Memory.Total))
			ch <- metric

			metric = m.memoryUsed.With(lbs)
			metric.Set(float64(info.Memory.Used))
			ch <- metric

			metric = m.memoryCached.With(lbs)
			metric.Set(float64(info.Memory.Cached))
			ch <- metric

			metric = m.swapTotal.With(lbs)
			metric.Set(float64(info.Swap.Total))
			ch <- metric

			metric = m.swapUsed.With(lbs)
			metric.Set(float64(info.Swap.Used))
			ch <- metric

			metric = m.swapFree.With(lbs)
			metric.Set(float64(info.Swap.Total - info.Swap.Used))
			ch <- metric

			metric = m.containersCount.With(lbs)
			metric.Set(float64(info.Docker.NumRunningContainers))
			ch <- metric

			for _, devinfo := range info.Disks {
				lbs["dev_name"] = devinfo.DevName
				lbs["mount_at"] = devinfo.MountAt

				metric = m.diskTotal.With(lbs)
				metric.Set(float64(devinfo.Total))
				ch <- metric

				metric = m.diskUsed.With(lbs)
				metric.Set(float64(devinfo.Used))
				ch <- metric

				metric = m.diskFree.With(lbs)
				metric.Set(float64(devinfo.Free))
				ch <- metric

				metric = m.diskINode.With(lbs)
				metric.Set(float64(devinfo.Inode))
				ch <- metric

				metric = m.diskIFree.With(lbs)
				metric.Set(float64(devinfo.Ifree))
				ch <- metric
			}

			// clean up additional label pairs
			delete(lbs, "dev_name")
			delete(lbs, "mount_at")

			for _, netinfo := range info.Traffics {
				lbs["device"] = netinfo.Name

				metric = m.networkRx.With(lbs)
				metric.Set(float64(netinfo.RxBytes))
				ch <- metric

				metric = m.networkTx.With(lbs)
				metric.Set(float64(netinfo.TxBytes))
				ch <- metric
			}

			for _, ioinfo := range info.DisksIO {
				lbs["device"] = ioinfo.DevName

				metric = m.diskIORead.With(lbs)
				metric.Set(float64(ioinfo.ReadBytes))
				ch <- metric

				metric = m.diskIOWrite.With(lbs)
				metric.Set(float64(ioinfo.WriteBytes))
				ch <- metric
			}
		}(node)
	}
	wg.Wait()
}
