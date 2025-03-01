package main

import (
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/minhu/prometheus-ethtool-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/safchain/ethtool"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

var (
	listenAddress = flag.String("web.listen-address", ":9417", "Address on which to expose metrics")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
	interfaces    = flag.String("interfaces", "", "Comma-separated list of interfaces to monitor (default: all interfaces)")
)

// getNetworkInterfaces returns a list of all available network interfaces with supported drivers
func getNetworkInterfaces() ([]string, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	// Initialize ethtool
	eth, err := ethtool.NewEthtool()
	if err != nil {
		return nil, err
	}
	defer eth.Close()

	var interfaces []string
	for _, link := range links {
		// Skip loopback
		if link.Attrs().Name == "lo" {
			continue
		}

		// Get driver info
		info, err := eth.DriverInfo(link.Attrs().Name)
		if err != nil {
			log.Debugf("Failed to get driver info for interface %s: %v", link.Attrs().Name, err)
			continue
		}

		// Only include mlx5 and ena interfaces
		if info.Driver == "mlx5_core" || info.Driver == "ena" {
			interfaces = append(interfaces, link.Attrs().Name)
			log.Debugf("Found supported interface %s with driver %s", link.Attrs().Name, info.Driver)
		} else {
			log.Debugf("Skipping unsupported interface %s with driver %s", link.Attrs().Name, info.Driver)
		}
	}
	return interfaces, nil
}

func main() {
	flag.Parse()

	// Configure logging
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	// Check if running as root
	if os.Geteuid() != 0 {
		log.Fatal("This exporter needs to be run with root privileges to access network statistics")
	}

	// Parse interfaces
	var ifaceList []string
	if *interfaces != "" {
		// When interfaces are manually specified, we'll still filter them
		eth, err := ethtool.NewEthtool()
		if err != nil {
			log.Fatalf("Failed to initialize ethtool: %v", err)
		}
		defer eth.Close()

		for _, iface := range strings.Split(*interfaces, ",") {
			info, err := eth.DriverInfo(iface)
			if err != nil {
				log.Warnf("Failed to get driver info for interface %s: %v", iface, err)
				continue
			}
			if info.Driver == "mlx5_core" || info.Driver == "ena" {
				ifaceList = append(ifaceList, iface)
				log.Debugf("Added manually specified interface %s with driver %s", iface, info.Driver)
			} else {
				log.Warnf("Skipping manually specified interface %s: unsupported driver %s", iface, info.Driver)
			}
		}
	} else {
		var err error
		ifaceList, err = getNetworkInterfaces()
		if err != nil {
			log.Fatalf("Failed to auto-detect network interfaces: %v", err)
		}
	}

	if len(ifaceList) == 0 {
		log.Fatal("No supported network interfaces found (only mlx5 and ena drivers are supported)")
	}

	// Create and register collector
	collector, err := collector.NewEthtoolCollector(ifaceList)
	if err != nil {
		log.Fatalf("Failed to create collector: %v", err)
	}
	defer collector.Close()

	prometheus.MustRegister(collector)

	// Setup HTTP server
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(`<html>
			<head><title>Network Interface Statistics Exporter</title></head>
			<body>
			<h1>Network Interface Statistics Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			<p>Only monitoring interfaces with mlx5 and ena drivers.</p>
			</body>
			</html>`)); err != nil {
			log.Errorf("Error writing response: %v", err)
		}
	})

	// Start server
	log.Infof("Starting network interface statistics exporter on %s", *listenAddress)
	log.Infof("Monitoring supported interfaces (mlx5/ena): %s", strings.Join(ifaceList, ", "))
	srv := &http.Server{
		Addr:         *listenAddress,
		Handler:      nil, // Use default handler
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Error starting HTTP server: %s", err)
	}
}
