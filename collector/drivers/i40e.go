package drivers

import (
	"regexp"
	"strconv"
)

// I40EMetricMapping defines which source metrics contribute to each basic metric.
var I40EMetricMapping = map[string][]string{
	CounterRxPackets: {"rx_packets"},
	CounterRxBytes:   {"rx_bytes"},
	CounterTxPackets: {"tx_packets"},
	CounterTxBytes:   {"tx_bytes"},
	CounterRxDrops:   {"rx_dropped", "rx_missed_errors"},
	CounterTxDrops:   {"tx_dropped"},
}

// I40EPhyMetricMapping defines which source metrics contribute to each physical metric.
var I40EPhyMetricMapping = map[string][]string{
	CounterRxPackets:   {"port.rx_unicast", "port.rx_multicast", "port.rx_broadcast"},
	CounterRxBytes:     {"port.rx_bytes"},
	CounterTxPackets:   {"port.tx_unicast", "port.tx_multicast", "port.tx_broadcast"},
	CounterTxBytes:     {"port.tx_bytes"},
	CounterRxDiscards:  {"port.rx_discards"},
	CounterTxDiscards:  {"port.tx_dropped_link_down"},
	CounterRxPauseCtrl: {"port.link_xon_rx", "port.link_xoff_rx"},
	CounterTxPauseCtrl: {"port.link_xon_tx", "port.link_xoff_tx"},
}

var i40eQueuePattern = regexp.MustCompile(`^(rx|tx)-(\d+)\.(packets|bytes)$`)

func processI40EBasicStats(rawStats map[string]uint64, stats *BasicStats) {
	for basicMetric, sourceMetrics := range I40EMetricMapping {
		total := sumMetrics(rawStats, sourceMetrics)

		switch basicMetric {
		case CounterRxPackets:
			stats.RxPackets = total
		case CounterRxBytes:
			stats.RxBytes = total
		case CounterTxPackets:
			stats.TxPackets = total
		case CounterTxBytes:
			stats.TxBytes = total
		case CounterRxDrops:
			stats.RxDrops = total
		case CounterTxDrops:
			stats.TxDrops = total
		}
	}
}

func processI40EPhyStats(rawStats map[string]uint64, phyStats *PhyStats) {
	for phyMetric, sourceMetrics := range I40EPhyMetricMapping {
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

func processI40EQueueStats(rawStats map[string]uint64) []QueueStats {
	queueMap := make(map[int]*QueueStats)

	for name, value := range rawStats {
		matches := i40eQueuePattern.FindStringSubmatch(name)
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

func processI40EStats(rawStats map[string]uint64) ProcessedStats {
	result := ProcessedStats{
		DriverSpecific: make(map[string]uint64),
	}

	processI40EBasicStats(rawStats, &result.Basic)

	result.Physical = &PhyStats{}
	processI40EPhyStats(rawStats, result.Physical)

	result.PerQueue = processI40EQueueStats(rawStats)

	for name, value := range rawStats {
		result.DriverSpecific["raw_"+name] = value
	}

	return result
}

func sumMetrics(stats map[string]uint64, names []string) uint64 {
	var total uint64
	for _, name := range names {
		total += stats[name]
	}
	return total
}
