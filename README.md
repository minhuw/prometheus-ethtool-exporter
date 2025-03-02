# Prometheus Ethtool Exporter

A Prometheus exporter that exposes NIC statistics obtained through `ethtool -S`.

## Features

- Provides standardized metrics across all supported drivers
- Supports per-queue statistics when available
- Auto-detects supported network interfaces
- Provides detailed network interface metrics for monitoring

## Prerequisites

- Go 1.21 or higher
- One or more Mellanox ConnectX network cards using the mlx5 driver

## Usage

```bash
# Run with default settings (port 9417)
sudo ./prometheus-ethtool-exporter

# Specify custom port
sudo ./prometheus-ethtool-exporter -port 9100

# Specify specific interfaces to monitor
sudo ./prometheus-ethtool-exporter -interfaces eth0,eth1
```

## Deployment

### Installation

```bash
go install github.com/minhu/prometheus-ethtool-exporter@latest
```

Or build from source:

```bash
git clone https://github.com/minhu/prometheus-ethtool-exporter.git
cd prometheus-ethtool-exporter
go build
```

### NixOS Deployment

1. Add the flake to your NixOS configuration:

```nix
{
  inputs.ethtool-exporter.url = "github:yourusername/prometheus-ethtool-exporter";
  
  outputs = { self, nixpkgs, ethtool-exporter, ... }: {
    nixosConfigurations.yourhostname = nixpkgs.lib.nixosSystem {
      modules = [
        ethtool-exporter.nixosModules.default
        # ...
      ];
    };
  };
}
```

2. Enable and configure the service in your `configuration.nix`:

```nix
{
  services.prometheus.exporters.ethtool = {
    enable = true;
    port = 9417;
    interfaces = [ "eth0" "eth1" ];
    openFirewall = true;
  };
}
```

### Docker Module

1. Build the Docker image:

```bash
docker build -t prometheus-ethtool-exporter .
```

2. Run the container:

```bash
docker run -d \
  --name ethtool-exporter \
  --net="host" \
  --cap-add=NET_ADMIN \
  --cap-add=NET_RAW \
  -p 9417:9417 \
  prometheus-ethtool-exporter
```

Note: The container requires `NET_ADMIN` and `NET_RAW` capabilities to access network interface statistics.

## Metrics

All metrics are exposed on `/metrics` endpoint. The exporter provides several types of metrics:

### Metrics Overview

#### Basic Network Metrics
| Metric Name | Type | Description |
|------------|------|-------------|
| `nic_rx_packets` | Counter | Total number of packets received |
| `nic_rx_bytes` | Counter | Total number of bytes received |
| `nic_rx_drops` | Counter | Total number of received packets dropped |
| `nic_tx_packets` | Counter | Total number of packets transmitted |
| `nic_tx_bytes` | Counter | Total number of bytes transmitted |
| `nic_tx_drops` | Counter | Total number of transmitted packets dropped |

#### Queue-Specific Metrics
| Metric Name | Type | Description |
|------------|------|-------------|
| `nic_queue_rx_packets` | Counter | Number of packets received on specific queue |
| `nic_queue_rx_bytes` | Counter | Number of bytes received on specific queue |
| `nic_queue_rx_drops` | Counter | Number of packets dropped on receive queue |
| `nic_queue_tx_packets` | Counter | Number of packets transmitted on specific queue |
| `nic_queue_tx_bytes` | Counter | Number of bytes transmitted on specific queue |
| `nic_queue_tx_drops` | Counter | Number of packets dropped on transmit queue |

#### Physical Layer Metrics
| Metric Name | Type | Description |
|------------|------|-------------|
| `nic_phy_rx_bytes` | Counter | Number of bytes received at physical layer |
| `nic_phy_tx_bytes` | Counter | Number of bytes transmitted at physical layer |
| `nic_phy_rx_packets` | Counter | Number of packets received at physical layer |
| `nic_phy_tx_packets` | Counter | Number of packets transmitted at physical layer |
| `nic_phy_rx_discards` | Counter | Number of packets discarded at physical layer receive |
| `nic_phy_tx_discards` | Counter | Number of packets discarded at physical layer transmit |
| `nic_phy_rx_pause_ctrl` | Counter | Number of pause control frames received |
| `nic_phy_tx_pause_ctrl` | Counter | Number of pause control frames transmitted |

#### Information Metrics
| Metric Name | Type | Description |
|------------|------|-------------|
| `nic_info` | Gauge | Network interface information (constant 1) |

## License

MIT License 