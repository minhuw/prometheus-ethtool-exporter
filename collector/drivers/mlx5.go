package drivers

import (
	"regexp"
	"strconv"
	"strings"
)

// Define constants for counter names
const (
	CounterRxPackets = "rx_packets"
	CounterRxBytes   = "rx_bytes"
	CounterTxPackets = "tx_packets"
	CounterTxBytes   = "tx_bytes"
	CounterRxDrops   = "rx_drops"
	CounterTxDrops   = "tx_drops"

	// Physical metrics
	CounterRxDiscards  = "rx_discards"
	CounterTxDiscards  = "tx_discards"
	CounterRxPauseCtrl = "rx_pause_ctrl"
	CounterTxPauseCtrl = "tx_pause_ctrl"
)

// MLX5MetricMapping defines which source metrics contribute to each basic metric
var MLX5MetricMapping = map[string][]string{
	CounterRxPackets: {"rx_packets"},
	CounterRxBytes:   {"rx_bytes"},
	CounterTxPackets: {"tx_packets"},
	CounterTxBytes:   {"tx_bytes"},
	CounterRxDrops:   {"rx_out_of_buffer", "rx_buff_alloc_err", "rx_wqe_err"},
	CounterTxDrops:   {"tx_cqe_err"},
}

// MLX5PhyMetricMapping defines which source metrics contribute to each physical metric
var MLX5PhyMetricMapping = map[string][]string{
	CounterRxPackets:   {"rx_packets_phy"},
	CounterRxBytes:     {"rx_bytes_phy"},
	CounterTxPackets:   {"tx_packets_phy"},
	CounterTxBytes:     {"tx_bytes_phy"},
	CounterRxDiscards:  {"rx_discards_phy"},
	CounterTxDiscards:  {"tx_discards_phy"},
	CounterRxPauseCtrl: {"rx_pause_ctrl_phy"},
	CounterTxPauseCtrl: {"tx_pause_ctrl_phy"},
}

// MLX5QueueMetricMapping defines which source metrics contribute to each queue metric
var MLX5QueueMetricMapping = map[string][]string{
	CounterRxPackets: {"rx_packets"},
	CounterRxBytes:   {"rx_bytes"},
	CounterTxPackets: {"tx_packets"},
	CounterTxBytes:   {"tx_bytes"},
	CounterRxDrops:   {"rx_wqe_err", "rx_buff_alloc_err"},
	CounterTxDrops:   {"tx_cqe_err"},
}

// Queue metric patterns
var (
	mlx5QueuePattern = regexp.MustCompile(`^(rx|tx)(\d+)_`)
)

// processMLX5BasicStats processes standard statistics for MLX5 NICs
func processMLX5BasicStats(rawStats map[string]uint64, stats *BasicStats) {
	// Process each basic metric using the mapping
	for basicMetric, sourceMetrics := range MLX5MetricMapping {
		var total uint64
		for _, sourceMetric := range sourceMetrics {
			if value, exists := rawStats[sourceMetric]; exists {
				total += value
			}
		}

		// Assign the total to the appropriate field in BasicStats
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

// processMLX5PhyStats processes physical layer statistics for MLX5 NICs
func processMLX5PhyStats(metrics map[string]uint64, phyStats *PhyStats) {
	// Process each physical metric using the mapping
	for phyMetric, sourceMetrics := range MLX5PhyMetricMapping {
		var total uint64
		for _, sourceMetric := range sourceMetrics {
			if value, exists := metrics[sourceMetric]; exists {
				total += value
			}
		}

		// Assign the total to the appropriate field in PhyStats
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

// processMLX5QueueStats processes per-queue statistics for MLX5 NICs
func processMLX5QueueStats(metrics map[string]uint64) []QueueStats {
	// Map to store queue stats by index
	queueMap := make(map[int]*QueueStats)

	// Process each metric
	for name, value := range metrics {
		// Extract queue index and base metric name
		if matches := mlx5QueuePattern.FindStringSubmatch(name); matches != nil {
			qIndex, err := strconv.Atoi(matches[2])
			if err != nil {
				continue
			}

			// Get or create queue stats
			qStats, exists := queueMap[qIndex]
			if !exists {
				queueMap[qIndex] = &QueueStats{
					QueueIndex: qIndex,
				}
				qStats = queueMap[qIndex]
			}

			// Process each queue metric using the mapping
			for queueMetric, sourceMetrics := range MLX5QueueMetricMapping {
				// Check if this metric matches any of the source metrics for this queue metric
				for _, sourceMetric := range sourceMetrics {
					// Construct the expected metric name with queue index
					// e.g., if sourceMetric is "rx_packets", and qIndex is 0,
					// expectedName would be "rx0_packets"
					direction := strings.Split(sourceMetric, "_")[0]          // "rx" or "tx"
					suffix := strings.TrimPrefix(sourceMetric, direction+"_") // "packets", "bytes", etc
					expectedName := direction + strconv.Itoa(qIndex) + "_" + suffix

					if name == expectedName {
						switch queueMetric {
						case CounterRxPackets:
							qStats.RxPackets = value
						case CounterRxBytes:
							qStats.RxBytes = value
						case CounterTxPackets:
							qStats.TxPackets = value
						case CounterTxBytes:
							qStats.TxBytes = value
						case CounterRxDrops:
							qStats.RxDrops += value
						case CounterTxDrops:
							qStats.TxDrops += value
						}
					}
				}
			}
		}
	}

	// Convert map to slice
	var result = make([]QueueStats, 0, len(queueMap))
	for _, stats := range queueMap {
		result = append(result, *stats)
	}

	return result
}

// processMLX5CustomStats processes MLX5-specific statistics
func processMLX5CustomStats(metrics map[string]uint64, driverSpecific map[string]uint64) {

}

func processMLX5Stats(rawStats map[string]uint64) ProcessedStats {
	result := ProcessedStats{
		DriverSpecific: make(map[string]uint64),
	}

	// Process basic metrics
	processMLX5BasicStats(rawStats, &result.Basic)

	// Process physical metrics
	result.Physical = &PhyStats{}
	processMLX5PhyStats(rawStats, result.Physical)

	// Process queue metrics
	result.PerQueue = processMLX5QueueStats(rawStats)

	// Process custom metrics and store raw metrics
	processMLX5CustomStats(rawStats, result.DriverSpecific)

	// Copy all raw metrics to DriverSpecific
	for name, value := range rawStats {
		result.DriverSpecific["raw_"+name] = value
	}

	return result
}
