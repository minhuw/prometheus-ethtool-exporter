package drivers

import (
	"regexp"
	"strconv"
)

// IXGBEMetricMapping defines which source metrics contribute to each basic metric.
var IXGBEMetricMapping = map[string][]string{
	CounterRxPackets: {"rx_packets"},
	CounterRxBytes:   {"rx_bytes"},
	CounterRxDrops: {
		"rx_dropped",
		"rx_missed_errors",
		"rx_no_buffer_count",
		"alloc_rx_page_failed",
		"alloc_rx_buff_failed",
		"rx_no_dma_resources",
	},
	CounterTxPackets: {"tx_packets"},
	CounterTxBytes:   {"tx_bytes"},
	CounterTxDrops:   {"tx_dropped"},
}

// IXGBEPhyMetricMapping defines which source metrics contribute to each physical metric.
var IXGBEPhyMetricMapping = map[string][]string{
	CounterRxPackets:   {"rx_pkts_nic"},
	CounterRxBytes:     {"rx_bytes_nic"},
	CounterTxPackets:   {"tx_pkts_nic"},
	CounterTxBytes:     {"tx_bytes_nic"},
	CounterRxDiscards:  {"rx_missed_errors", "rx_no_buffer_count", "rx_no_dma_resources"},
	CounterTxDiscards:  {"tx_dropped"},
	CounterRxPauseCtrl: {"rx_flow_control_xon", "rx_flow_control_xoff"},
	CounterTxPauseCtrl: {"tx_flow_control_xon", "tx_flow_control_xoff"},
}

var ixgbeQueuePattern = regexp.MustCompile(`^(rx|tx)_queue_(\d+)_(packets|bytes)$`)

func processIXGBEBasicStats(rawStats map[string]uint64, stats *BasicStats) {
	for basicMetric, sourceMetrics := range IXGBEMetricMapping {
		total := sumMetrics(rawStats, sourceMetrics)

		switch basicMetric {
		case CounterRxPackets:
			stats.RxPackets = total
		case CounterRxBytes:
			stats.RxBytes = total
		case CounterRxDrops:
			stats.RxDrops = total
		case CounterTxPackets:
			stats.TxPackets = total
		case CounterTxBytes:
			stats.TxBytes = total
		case CounterTxDrops:
			stats.TxDrops = total
		}
	}
}

func processIXGBEPhyStats(rawStats map[string]uint64, phyStats *PhyStats) {
	for phyMetric, sourceMetrics := range IXGBEPhyMetricMapping {
		total := sumMetrics(rawStats, sourceMetrics)

		switch phyMetric {
		case CounterRxPackets:
			phyStats.RxPackets = total
		case CounterRxBytes:
			phyStats.RxBytes = total
		case CounterTxPackets:
			phyStats.TxPackets = total
		case CounterTxBytes:
			phyStats.TxBytes = total
		case CounterRxDiscards:
			phyStats.RxDiscarded = total
		case CounterTxDiscards:
			phyStats.TxDiscarded = total
		case CounterRxPauseCtrl:
			phyStats.RxPauseCtrl = total
		case CounterTxPauseCtrl:
			phyStats.TxPauseCtrl = total
		}
	}
}

func processIXGBEQueueStats(rawStats map[string]uint64) []QueueStats {
	queueMap := make(map[int]*QueueStats)

	for name, value := range rawStats {
		matches := ixgbeQueuePattern.FindStringSubmatch(name)
		if matches == nil {
			continue
		}

		qIndex, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}

		qStats, exists := queueMap[qIndex]
		if !exists {
			qStats = &QueueStats{QueueIndex: qIndex}
			queueMap[qIndex] = qStats
		}

		switch {
		case matches[1] == "rx" && matches[3] == "packets":
			qStats.RxPackets = value
		case matches[1] == "rx" && matches[3] == "bytes":
			qStats.RxBytes = value
		case matches[1] == "tx" && matches[3] == "packets":
			qStats.TxPackets = value
		case matches[1] == "tx" && matches[3] == "bytes":
			qStats.TxBytes = value
		}
	}

	result := make([]QueueStats, 0, len(queueMap))
	for _, stats := range queueMap {
		result = append(result, *stats)
	}

	return result
}

func processIXGBEStats(rawStats map[string]uint64) ProcessedStats {
	result := ProcessedStats{
		DriverSpecific: make(map[string]uint64),
	}

	processIXGBEBasicStats(rawStats, &result.Basic)

	result.Physical = &PhyStats{}
	processIXGBEPhyStats(rawStats, result.Physical)

	result.PerQueue = processIXGBEQueueStats(rawStats)

	for name, value := range rawStats {
		result.DriverSpecific["raw_"+name] = value
	}

	return result
}
