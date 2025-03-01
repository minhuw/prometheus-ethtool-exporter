package drivers

// Known NIC drivers
const (
	DriverMLX5 = "mlx5_core"
)

// BasicStats contains the basic metrics every NIC should provide
type BasicStats struct {
	RxPackets uint64 // Total received packets
	RxBytes   uint64 // Total received bytes
	RxDrops   uint64 // Total dropped RX packets
	TxPackets uint64 // Total transmitted packets
	TxBytes   uint64 // Total transmitted bytes
	TxDrops   uint64 // Total dropped TX packets
}

// PhyStats contains physical layer statistics
type PhyStats struct {
	RxPackets   uint64 // Packets received at physical layer
	RxBytes     uint64 // Bytes received at physical layer
	TxPackets   uint64 // Packets transmitted at physical layer
	TxBytes     uint64 // Bytes transmitted at physical layer
	RxDiscarded uint64 // Packets discarded at physical layer
	TxDiscarded uint64 // Packets discarded at physical layer
	RxPauseCtrl uint64 // Received pause control frames
	TxPauseCtrl uint64 // Transmitted pause control frames
}

// QueueStats contains per-queue statistics
type QueueStats struct {
	QueueIndex int
	RxPackets  uint64
	RxBytes    uint64
	TxPackets  uint64
	TxBytes    uint64
	RxDrops    uint64
	TxDrops    uint64
}

// ProcessedStats contains both basic and driver-specific metrics
type ProcessedStats struct {
	Basic          BasicStats
	Physical       *PhyStats // Physical layer statistics, may be nil if not supported
	PerQueue       []QueueStats
	DriverSpecific map[string]uint64
}

// NICInfo contains information about a network interface
type NICInfo struct {
	Name       string
	Driver     string
	DriverType string
	Version    string
}

// GetMetricPrefix returns the metric prefix for a driver type
func GetMetricPrefix(driverType string) string {
	switch driverType {
	case DriverMLX5:
		return "mlx5"
	default:
		return "unknown"
	}
}

// ProcessDriverStats processes driver-specific statistics
func ProcessDriverStats(driverType string, rawStats map[string]uint64) ProcessedStats {
	switch driverType {
	case DriverMLX5:
		return processMLX5Stats(rawStats)
	}

	return ProcessedStats{
		DriverSpecific: make(map[string]uint64),
	}
}
