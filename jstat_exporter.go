package main

import (
	"flag"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/promlog"
	stdlog "log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"

	"github.com/go-kit/log"
)

const (
	namespace = "jstat"
)

var (
	listenAddress = flag.String("web.listen-address", ":9010", "Address on which to expose metrics and web interface.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	maxRequests   = flag.Int("web.max-requests", 40, "Maximum number of parallel scrape requests. Use 0 to disable.")
	jstatPath     = flag.String("jstat.path", "/usr/bin/jstat", "jstat path")
	targetPid     = flag.String("target.pid", ":0", "target pid")
)

type Exporter struct {
	jstatPath  string
	targetPid  string
	newMax     prometheus.Gauge
	newCommit  prometheus.Gauge
	oldMax     prometheus.Gauge
	oldCommit  prometheus.Gauge
	metaMax    prometheus.Gauge
	metaCommit prometheus.Gauge
	metaUsed   prometheus.Gauge
	oldUsed    prometheus.Gauge
	sv0Used    prometheus.Gauge
	sv1Used    prometheus.Gauge
	edenUsed   prometheus.Gauge
	fgcTimes   prometheus.Gauge
	fgcSec     prometheus.Gauge
	logger     log.Logger
}

func NewExporter(jstatPath string, targetPid string) *Exporter {
	return &Exporter{
		jstatPath: jstatPath,
		targetPid: targetPid,
		newMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "newMax",
			Help:      "newMax",
		}),
		newCommit: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "newCommit",
			Help:      "newCommit",
		}),
		oldMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "oldMax",
			Help:      "oldMax",
		}),
		oldCommit: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "oldCommit",
			Help:      "oldCommit",
		}),
		metaMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "metaMax",
			Help:      "metaMax",
		}),
		metaCommit: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "metaCommit",
			Help:      "metaCommit",
		}),
		metaUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "metaUsed",
			Help:      "metaUsed",
		}),
		oldUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "oldUsed",
			Help:      "oldUsed",
		}),
		sv0Used: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sv0Used",
			Help:      "sv0Used",
		}),
		sv1Used: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sv1Used",
			Help:      "sv1Used",
		}),
		edenUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "edenUsed",
			Help:      "edenUsed",
		}),
		fgcTimes: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "fgcTimes",
			Help:      "fgcTimes",
		}),
		fgcSec: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "fgcSec",
			Help:      "fgcSec",
		}),
	}
}

// Describe implements the prometheus.Collector interface.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.newMax.Describe(ch)
	e.newCommit.Describe(ch)
	e.oldMax.Describe(ch)
	e.oldCommit.Describe(ch)
	e.metaMax.Describe(ch)
	e.metaCommit.Describe(ch)
	e.metaUsed.Describe(ch)
	e.oldUsed.Describe(ch)
	e.sv0Used.Describe(ch)
	e.sv1Used.Describe(ch)
	e.edenUsed.Describe(ch)
	e.fgcTimes.Describe(ch)
	e.fgcSec.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.JstatGccapacity(ch)
	e.JstatGcold(ch)
	e.JstatGcnew(ch)
	e.JstatGc(ch)
}

func (e *Exporter) JstatGccapacity(ch chan<- prometheus.Metric) {
	out, err := exec.Command(e.jstatPath, "-gccapacity", e.targetPid).Output()
	if err != nil {
		stdlog.Fatal(err)
	}

	for i, line := range strings.Split(string(out), "\n") {
		if i == 1 {
			parts := strings.Fields(line)
			newMax, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				stdlog.Fatal(err)
			}
			e.newMax.Set(newMax)
			e.newMax.Collect(ch)
			newCommit, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				stdlog.Fatal(err)
			}
			e.newCommit.Set(newCommit)
			e.newCommit.Collect(ch)
			oldMax, err := strconv.ParseFloat(parts[7], 64)
			if err != nil {
				stdlog.Fatal(err)
			}
			e.oldMax.Set(oldMax)
			e.oldMax.Collect(ch)
			oldCommit, err := strconv.ParseFloat(parts[8], 64)
			if err != nil {
				stdlog.Fatal(err)
			}
			e.oldCommit.Set(oldCommit)
			e.oldCommit.Collect(ch)
			metaMax, err := strconv.ParseFloat(parts[11], 64)
			if err != nil {
				stdlog.Fatal(err)
			}
			e.metaMax.Set(metaMax)
			e.metaMax.Collect(ch)
			metaCommit, err := strconv.ParseFloat(parts[12], 64)
			if err != nil {
				stdlog.Fatal(err)
			}
			e.metaCommit.Set(metaCommit)
			e.metaCommit.Collect(ch)
		}
	}
}

func (e *Exporter) JstatGcold(ch chan<- prometheus.Metric) {
	out, err := exec.Command(e.jstatPath, "-gcold", e.targetPid).Output()
	if err != nil {
		stdlog.Fatal(err)
	}

	for i, line := range strings.Split(string(out), "\n") {
		if i == 1 {
			parts := strings.Fields(line)
			metaUsed, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				stdlog.Fatal(err)
			}
			e.metaUsed.Set(metaUsed) // MU: Metaspace utilization (kB).
			e.metaUsed.Collect(ch)
			oldUsed, err := strconv.ParseFloat(parts[5], 64)
			if err != nil {
				stdlog.Fatal(err)
			}
			e.oldUsed.Set(oldUsed) // OU: Old space utilization (kB).
			e.oldUsed.Collect(ch)
		}
	}
}

func (e *Exporter) JstatGcnew(ch chan<- prometheus.Metric) {
	out, err := exec.Command(e.jstatPath, "-gcnew", e.targetPid).Output()
	if err != nil {
		stdlog.Fatal(err)
	}

	for i, line := range strings.Split(string(out), "\n") {
		if i == 1 {
			parts := strings.Fields(line)
			sv0Used, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				stdlog.Fatal(err)
			}
			e.sv0Used.Set(sv0Used)
			e.sv0Used.Collect(ch)
			sv1Used, err := strconv.ParseFloat(parts[3], 64)
			if err != nil {
				stdlog.Fatal(err)
			}
			e.sv1Used.Set(sv1Used)
			e.sv1Used.Collect(ch)
			edenUsed, err := strconv.ParseFloat(parts[8], 64)
			if err != nil {
				stdlog.Fatal(err)
			}
			e.edenUsed.Set(edenUsed)
			e.edenUsed.Collect(ch)
		}
	}
}

func (e *Exporter) JstatGc(ch chan<- prometheus.Metric) {
	out, err := exec.Command(e.jstatPath, "-gc", e.targetPid).Output()
	if err != nil {
		stdlog.Fatal(err)
	}

	for i, line := range strings.Split(string(out), "\n") {
		if i == 1 {
			parts := strings.Fields(line)
			fgcTimes, err := strconv.ParseFloat(parts[14], 64)
			if err != nil {
				stdlog.Fatal(err)
			}
			e.fgcTimes.Set(fgcTimes)
			e.fgcTimes.Collect(ch)
			fgcSec, err := strconv.ParseFloat(parts[15], 64)
			if err != nil {
				stdlog.Fatal(err)
			}
			e.fgcSec.Set(fgcSec)
			e.fgcSec.Collect(ch)
		}
	}
}

func main() {
	flag.Parse()

	promlogConfig := &promlog.Config{}
	logger := promlog.New(promlogConfig)

	exporterMetricsRegistry := prometheus.NewRegistry()
	r := prometheus.NewRegistry()

	r.MustRegister(version.NewCollector("jstat_exporter"))

	exporter := NewExporter(*jstatPath, *targetPid)
	if err := r.Register(exporter); err != nil {
		stdlog.Fatalf("couldn't register jstat collector: %s", err)
	}

	level.Info(logger).Log("Starting Server: %s", *listenAddress)
	handler := promhttp.HandlerFor(
		prometheus.Gatherers{exporterMetricsRegistry, r},
		promhttp.HandlerOpts{
			ErrorLog:            stdlog.New(log.NewStdlibAdapter(level.Error(logger)), "", 0),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: *maxRequests,
		},
	)

	http.Handle(*metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>jstat Exporter</title></head>
		<body>
		<h1>jstat Exporter</h1>
		<p><a href="` + *metricsPath + `">Metrics</a></p>
		</body>
		</html>`))
	})
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		stdlog.Fatal(err)
	}
}
