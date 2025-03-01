package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/minhu/prometheus-ethtool-exporter/collector"
	"github.com/minhu/prometheus-ethtool-exporter/collector/drivers"
	"github.com/safchain/ethtool"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

var (
	iface    = flag.String("interface", "", "Network interface to inspect")
	format   = flag.String("format", "text", "Output format: text, json")
	verbose  = flag.Bool("verbose", false, "Show verbose output")
	listOnly = flag.Bool("list", false, "List available network interfaces and exit")
)

type debugInfo struct {
	Interface      string                  `json:"interface"`
	Driver         string                  `json:"driver"`
	DriverType     string                  `json:"driver_type"`
	Version        string                  `json:"version"`
	EthtoolStats   map[string]uint64       `json:"ethtool_stats,omitempty"`
	ProcessedStats *drivers.ProcessedStats `json:"processed_stats,omitempty"`
}

func main() {
	flag.Parse()

	if os.Geteuid() != 0 {
		log.Fatal("This tool needs to be run with root privileges")
	}

	// Configure logging
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	// List available interfaces if requested
	if *listOnly {
		listInterfaces()
		return
	}

	// Validate interface argument
	if *iface == "" {
		log.Fatal("Please specify a network interface with -interface")
	}

	// Get interface information
	info, err := getDebugInfo(*iface)
	if err != nil {
		log.Fatalf("Failed to get interface information: %v", err)
	}

	// Output results
	switch strings.ToLower(*format) {
	case "json":
		outputJSON(info)
	default:
		outputText(info)
	}
}

func getDebugInfo(ifaceName string) (*debugInfo, error) {
	// Get interface
	link, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface: %v", err)
	}

	// Create ethtool instance
	eth, err := ethtool.NewEthtool()
	if err != nil {
		return nil, fmt.Errorf("failed to create ethtool: %v", err)
	}
	defer eth.Close()

	// Get NIC information
	nicInfo, err := collector.GetNICInfo(eth, link)
	if err != nil {
		return nil, fmt.Errorf("failed to get NIC info: %v", err)
	}

	info := &debugInfo{
		Interface:  ifaceName,
		Driver:     nicInfo.Driver,
		DriverType: nicInfo.DriverType,
		Version:    nicInfo.Version,
	}

	// Get ethtool statistics
	if stats, err := eth.Stats(ifaceName); err == nil {
		info.EthtoolStats = stats
	}

	// Get processed statistics
	if info.EthtoolStats != nil {
		processed := drivers.ProcessDriverStats(info.DriverType, info.EthtoolStats)
		info.ProcessedStats = &processed
	}

	return info, nil
}

func listInterfaces() {
	links, err := netlink.LinkList()
	if err != nil {
		log.Fatalf("Failed to list interfaces: %v", err)
	}

	// Create ethtool instance
	eth, err := ethtool.NewEthtool()
	if err != nil {
		log.Fatalf("Failed to create ethtool: %v", err)
	}
	defer eth.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Interface\tDriver\tType\tState")
	fmt.Fprintln(w, "---------\t------\t----\t-----")

	for _, link := range links {
		attrs := link.Attrs()
		if attrs.Name == "lo" {
			continue // Skip loopback
		}

		nicInfo, err := collector.GetNICInfo(eth, link)
		driver := "unknown"
		driverType := "unknown"
		if err == nil {
			driver = nicInfo.Driver
			driverType = nicInfo.DriverType
		}

		state := "DOWN"
		if attrs.OperState == netlink.OperUp {
			state = "UP"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", attrs.Name, driver, driverType, state)
	}
	w.Flush()
}

func outputJSON(info *debugInfo) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(info); err != nil {
		log.Fatalf("Failed to output JSON: %v", err)
	}
}

func outputText(info *debugInfo) {
	fmt.Printf("Interface Information:\n")
	fmt.Printf("  Name:        %s\n", info.Interface)
	fmt.Printf("  Driver:      %s\n", info.Driver)
	fmt.Printf("  Driver Type: %s\n", info.DriverType)
	fmt.Printf("  Version:     %s\n", info.Version)

	if info.ProcessedStats != nil {
		fmt.Printf("\nStandard Statistics:\n")
		standardStats := map[string]uint64{
			"rx_packets": info.ProcessedStats.Basic.RxPackets,
			"rx_bytes":   info.ProcessedStats.Basic.RxBytes,
			"rx_drops":   info.ProcessedStats.Basic.RxDrops,
			"tx_packets": info.ProcessedStats.Basic.TxPackets,
			"tx_bytes":   info.ProcessedStats.Basic.TxBytes,
			"tx_drops":   info.ProcessedStats.Basic.TxDrops,
		}
		printStats(standardStats)

		if info.ProcessedStats.Physical != nil {
			fmt.Printf("\nPhysical Layer Statistics:\n")
			phyStats := map[string]uint64{
				"rx_packets":    info.ProcessedStats.Physical.RxPackets,
				"rx_bytes":      info.ProcessedStats.Physical.RxBytes,
				"tx_packets":    info.ProcessedStats.Physical.TxPackets,
				"tx_bytes":      info.ProcessedStats.Physical.TxBytes,
				"rx_discarded":  info.ProcessedStats.Physical.RxDiscarded,
				"tx_discarded":  info.ProcessedStats.Physical.TxDiscarded,
				"rx_pause_ctrl": info.ProcessedStats.Physical.RxPauseCtrl,
				"tx_pause_ctrl": info.ProcessedStats.Physical.TxPauseCtrl,
			}
			printStats(phyStats)
		}

		if len(info.ProcessedStats.PerQueue) > 0 {
			fmt.Printf("\nPer-Queue Statistics:\n")
			for _, q := range info.ProcessedStats.PerQueue {
				fmt.Printf("  Queue %d:\n", q.QueueIndex)
				queueStats := map[string]uint64{
					"rx_packets": q.RxPackets,
					"rx_bytes":   q.RxBytes,
					"rx_drops":   q.RxDrops,
					"tx_packets": q.TxPackets,
					"tx_bytes":   q.TxBytes,
					"tx_drops":   q.TxDrops,
				}
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				for k, v := range queueStats {
					fmt.Fprintf(w, "    %s:\t%d\n", k, v)
				}
				w.Flush()
			}
		}

		if len(info.ProcessedStats.DriverSpecific) > 0 {
			fmt.Printf("\nDriver-Specific Statistics:\n")
			printStats(info.ProcessedStats.DriverSpecific)
		}
	}

	if *verbose && len(info.EthtoolStats) > 0 {
		fmt.Printf("\nRaw Ethtool Statistics:\n")
		printStats(info.EthtoolStats)
	}
}

func printStats(stats map[string]uint64) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Sort keys for consistent output
	keys := make([]string, 0, len(stats))
	for k := range stats {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Print stats
	for _, k := range keys {
		fmt.Fprintf(w, "  %s:\t%d\n", k, stats[k])
	}
	w.Flush()
}
