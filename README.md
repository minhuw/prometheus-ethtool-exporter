# Prometheus Ethtool Exporter

A Prometheus exporter that exposes NIC statistics obtained through `ethtool -S`.

## Features

- Provides standardized metrics across all supported drivers
- Supports per-queue statistics when available
- Auto-detects supported network interfaces
- Provides detailed network interface metrics for monitoring

## Prerequisites

- Go 1.21 or higher
- Root privileges (required for netlink access)
- One or more supported network interfaces

## Installation

```bash
go install github.com/minhu/prometheus-ethtool-exporter@latest
```

Or build from source:

```bash
git clone https://github.com/minhu/prometheus-ethtool-exporter.git
cd prometheus-ethtool-exporter
go build
```

## Usage

```bash
# Run with default settings (port 9417)
sudo ./prometheus-ethtool-exporter

# Specify custom port
sudo ./prometheus-ethtool-exporter -port 9100

# Specify specific interfaces to monitor
sudo ./prometheus-ethtool-exporter -interfaces eth0,eth1
```

## Metrics

All metrics are exposed on `/metrics` endpoint. The exporter provides several types of metrics:

### Standard Metrics

These metrics are standardized across all supported drivers and have the prefix `nic_standard_`:

```
# HELP nic_standard_rx_packets_total Standardized network interface statistic
# TYPE nic_standard_rx_packets_total counter
nic_standard_rx_packets_total{interface="eth0",driver="ena"} 1234567

# Standard metrics available:
- rx_packets_total: Total received packets
- rx_bytes_total: Total received bytes
- tx_packets_total: Total transmitted packets
- tx_bytes_total: Total transmitted bytes
- rx_missed_total: Packets missed due to lack of descriptors
- rx_discarded_total: Packets discarded due to buffer exhaustion
- tx_errors_total: Total transmit errors
```

### Per-Queue Metrics

When available, per-queue statistics are exposed with the prefix `nic_queue_` and include a queue label:

```
# HELP nic_queue_rx_packets_total Per-queue network interface statistic
# TYPE nic_queue_rx_packets_total counter
nic_queue_rx_packets_total{interface="eth0",driver="mlx5",queue="0"} 123456

# Per-queue metrics available:
- queue_rx_packets_total: Received packets per queue
- queue_rx_bytes_total: Received bytes per queue
- queue_tx_packets_total: Transmitted packets per queue
- queue_tx_bytes_total: Transmitted bytes per queue
- queue_rx_missed_total: Missed packets per queue
- queue_rx_discarded_total: Discarded packets per queue
- queue_tx_errors_total: Transmit errors per queue
```

### Driver-Specific Metrics

#### Mellanox ConnectX (mlx5)

Metrics with prefix `nic_mlx5_`:
- rx_out_of_buffer_total
- rx_cqe_error_total
- rx_wqe_error_total
- tx_queue_stopped_total
- rx_bytes_total
- tx_bytes_total
- rx_packets_total
- tx_packets_total

#### Amazon ENA

Metrics with prefix `nic_ena_`:
- rx_drops_total
- tx_drops_total
- rx_overruns_total
- tx_timeout_total
- rx_bw_allowance_exceeded_total
- tx_bw_allowance_exceeded_total
- pps_allowance_exceeded_total

#### VirtIO Network

Metrics with prefix `nic_virtio_`:
- rx_queue_full_total
- tx_timeout_total
- rx_alloc_failed_total
- tx_busy_total

#### Broadcom Tigon3 (tg3)

Metrics with prefix `nic_tg3_`:
- rx_errors_total
- tx_errors_total
- rx_crc_errors_total
- rx_frame_errors_total
- rx_missed_total
- tx_aborted_total
- tx_carrier_errors_total
- dma_read_failed_total
- dma_write_failed_total
- mac_regs_failed_total

### Interface Information

```
# HELP nic_info Network interface information
# TYPE nic_info gauge
nic_info{interface="eth0",driver="ena",version="1"} 1
```

## Labels

All metrics include the following labels:
- `interface`: Network interface name
- `driver`: Driver type (mlx5, ena, virtio, or tg3)
- `queue`: Queue number (only for per-queue metrics)

## License

MIT License 