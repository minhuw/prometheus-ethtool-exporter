package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/safchain/ethtool"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"

	"github.com/minhu/prometheus-ethtool-exporter/collector/drivers"
)

// EthtoolCollector implements the prometheus.Collector interface.
type EthtoolCollector struct {
	interfaces []string
	metrics    map[string]*prometheus.Desc
	ethtool    *ethtool.Ethtool
}

// NewEthtoolCollector creates a new collector for the specified interfaces.
func NewEthtoolCollector(interfaces []string) (*EthtoolCollector, error) {
	eth, err := ethtool.NewEthtool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ethtool: %v", err)
	}

	return &EthtoolCollector{
		interfaces: interfaces,
		metrics:    make(map[string]*prometheus.Desc),
		ethtool:    eth,
	}, nil
}

// Close releases resources used by the collector.
func (c *EthtoolCollector) Close() error {
	if c.ethtool != nil {
		c.ethtool.Close()
	}
	return nil
}

// Describe implements prometheus.Collector.
func (c *EthtoolCollector) Describe(ch chan<- *prometheus.Desc) {
	// Since we don't know the metrics beforehand, we'll create them dynamically
	// during collection. This is a no-op.
}

// Collect implements prometheus.Collector.
func (c *EthtoolCollector) Collect(ch chan<- prometheus.Metric) {
	for _, ifaceName := range c.interfaces {
		// Get interface information
		link, err := netlink.LinkByName(ifaceName)
		if err != nil {
			log.Errorf("Failed to get interface %s: %v", ifaceName, err)
			continue
		}

		// Get NIC information and check if it's supported
		nicInfo, err := GetNICInfo(c.ethtool, link)
		if err != nil {
			log.Debugf("Skipping interface %s: %v", ifaceName, err)
			continue
		}

		// Add driver type label
		labels := []string{"interface", "driver"}
		labelValues := []string{ifaceName, nicInfo.DriverType}

		// Collect driver-specific statistics
		ethtoolStats, err := c.getEthtoolStats(ifaceName)
		if err != nil {
			log.Debugf("Failed to collect ethtool stats for interface %s: %v", ifaceName, err)
			continue
		}

		// Process all statistics
		processedStats := drivers.ProcessDriverStats(nicInfo.DriverType, ethtoolStats)

		// Export basic metrics
		basicMetrics := map[string]uint64{
			"rx_packets": processedStats.Basic.RxPackets,
			"rx_bytes":   processedStats.Basic.RxBytes,
			"rx_drops":   processedStats.Basic.RxDrops,
			"tx_packets": processedStats.Basic.TxPackets,
			"tx_bytes":   processedStats.Basic.TxBytes,
			"tx_drops":   processedStats.Basic.TxDrops,
		}

		for name, value := range basicMetrics {
			desc := c.getOrCreateMetricDesc(
				name,
				"Basic network interface statistic",
				labels,
			)
			metric := prometheus.MustNewConstMetric(
				desc,
				prometheus.CounterValue,
				float64(value),
				labelValues...,
			)
			ch <- metric
		}

		// Export per-queue metrics
		queueLabels := append(labels, "queue")
		for _, qStats := range processedStats.PerQueue {
			queueLabelValues := append(labelValues, fmt.Sprintf("%d", qStats.QueueIndex))

			queueMetrics := map[string]uint64{
				"queue_rx_packets": qStats.RxPackets,
				"queue_rx_bytes":   qStats.RxBytes,
				"queue_rx_drops":   qStats.RxDrops,
				"queue_tx_packets": qStats.TxPackets,
				"queue_tx_bytes":   qStats.TxBytes,
				"queue_tx_drops":   qStats.TxDrops,
			}

			for name, value := range queueMetrics {
				desc := c.getOrCreateMetricDesc(
					name,
					"Per-queue network interface statistic",
					queueLabels,
				)
				metric := prometheus.MustNewConstMetric(
					desc,
					prometheus.CounterValue,
					float64(value),
					queueLabelValues...,
				)
				ch <- metric
			}
		}

		// Add PHY statistics
		if processedStats.Physical != nil {
			phyStats := map[string]uint64{
				"rx_bytes":      processedStats.Physical.RxBytes,
				"tx_bytes":      processedStats.Physical.TxBytes,
				"rx_packets":    processedStats.Physical.RxPackets,
				"tx_packets":    processedStats.Physical.TxPackets,
				"rx_discards":   processedStats.Physical.RxDiscarded,
				"tx_discards":   processedStats.Physical.TxDiscarded,
				"rx_pause_ctrl": processedStats.Physical.RxPauseCtrl,
				"tx_pause_ctrl": processedStats.Physical.TxPauseCtrl,
			}

			for name, value := range phyStats {
				desc := c.getOrCreateMetricDesc(
					"phy_"+name,
					"PHY drops for network interface",
					labels,
				)
				ch <- prometheus.MustNewConstMetric(
					desc,
					prometheus.CounterValue,
					float64(value),
					labelValues...,
				)
			}
		}

		// Add driver info metric
		infoDesc := c.getOrCreateMetricDesc(
			"info",
			"Network interface information",
			append(labels, "version"),
		)
		ch <- prometheus.MustNewConstMetric(
			infoDesc,
			prometheus.GaugeValue,
			1,
			append(labelValues, nicInfo.Version)...,
		)
	}
}

// getOrCreateMetricDesc creates or returns an existing metric description.
func (c *EthtoolCollector) getOrCreateMetricDesc(name, help string, labels []string) *prometheus.Desc {
	if desc, exists := c.metrics[name]; exists {
		return desc
	}

	desc := prometheus.NewDesc(
		prometheus.BuildFQName("nic", "", name),
		help,
		labels,
		nil,
	)
	c.metrics[name] = desc
	return desc
}

// getEthtoolStats retrieves NIC-specific statistics using netlink ethtool interface.
func (c *EthtoolCollector) getEthtoolStats(iface string) (map[string]uint64, error) {
	stats, err := c.ethtool.Stats(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to get ethtool stats: %v", err)
	}
	return stats, nil
}

// GetNICInfo retrieves information about a network interface.
func GetNICInfo(eth *ethtool.Ethtool, link netlink.Link) (*drivers.NICInfo, error) {
	info, err := eth.DriverInfo(link.Attrs().Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver info: %v", err)
	}

	// Only accept mlx5 and ena drivers
	if info.Driver != "mlx5_core" && info.Driver != "ena" {
		return nil, fmt.Errorf("unsupported driver: %s (only mlx5 and ena drivers are supported)", info.Driver)
	}

	return &drivers.NICInfo{
		DriverType: info.Driver,
		Version:    info.Version,
	}, nil
}
