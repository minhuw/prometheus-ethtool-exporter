package drivers

import (
	"regexp"
	"strconv"
)

// ICEMetricMapping defines which source metrics contribute to each basic metric.
var ICEMetricMapping = map[string][]string{
	CounterRxPackets: {"rx_unicast", "rx_multicast", "rx_broadcast"},
	CounterRxBytes:   {"rx_bytes"},
	CounterRxDrops:   {"rx_dropped", "rx_alloc_fail", "rx_pg_alloc_fail"},
	CounterTxPackets: {"tx_unicast", "tx_multicast", "tx_broadcast"},
	CounterTxBytes:   {"tx_bytes"},
	CounterTxDrops:   {"tx_errors", "tx_dropped_link_down.nic"},
}

// ICEPhyMetricMapping defines which source metrics contribute to each physical metric.
var ICEPhyMetricMapping = map[string][]string{
	CounterRxPackets:   {"rx_unicast.nic", "rx_multicast.nic", "rx_broadcast.nic"},
	CounterRxBytes:     {"rx_bytes.nic"},
	CounterTxPackets:   {"tx_unicast.nic", "tx_multicast.nic", "tx_broadcast.nic"},
	CounterTxBytes:     {"tx_bytes.nic"},
	CounterRxDiscards:  {"rx_dropped.nic"},
	CounterTxDiscards:  {"tx_dropped_link_down.nic"},
	CounterRxPauseCtrl: {"link_xon_rx.nic", "link_xoff_rx.nic"},
	CounterTxPauseCtrl: {"link_xon_tx.nic", "link_xoff_tx.nic"},
}

var iceQueuePattern = regexp.MustCompile(`^(rx|tx)_queue_(\d+)_(packets|bytes)$`)

func processICEBasicStats(rawStats map[string]uint64, stats *BasicStats) {
	for basicMetric, sourceMetrics := range ICEMetricMapping {
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

func processICEPhyStats(rawStats map[string]uint64, phyStats *PhyStats) {
	for phyMetric, sourceMetrics := range ICEPhyMetricMapping {
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

func processICEQueueStats(rawStats map[string]uint64) []QueueStats {
	queueMap := make(map[int]*QueueStats)

	for name, value := range rawStats {
		matches := iceQueuePattern.FindStringSubmatch(name)
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

func processICEStats(rawStats map[string]uint64) ProcessedStats {
	result := ProcessedStats{
		DriverSpecific: make(map[string]uint64),
	}

	processICEBasicStats(rawStats, &result.Basic)

	result.Physical = &PhyStats{}
	processICEPhyStats(rawStats, result.Physical)

	result.PerQueue = processICEQueueStats(rawStats)

	for name, value := range rawStats {
		result.DriverSpecific["raw_"+name] = value
	}

	return result
}
