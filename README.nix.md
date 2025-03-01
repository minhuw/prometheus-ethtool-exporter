# NixOS Integration

This exporter can be integrated into your NixOS configuration using the provided flake.

## Quick Start

1. Add the flake to your NixOS configuration:

```nix
{
  inputs.ethtool-exporter.url = "github:yourusername/prometheus-ethtool-exporter";
  
  outputs = { self, nixpkgs, ethtool-exporter, ... }: {
    nixosConfigurations.yourhostname = nixpkgs.lib.nixosSystem {
      # ...
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

## Options

- `enable`: Enable the ethtool exporter
- `port`: Port to listen on (default: 9417)
- `interfaces`: List of network interfaces to monitor
- `openFirewall`: Whether to open the port in the firewall
- `capabilities`: Required Linux capabilities (default: ["CAP_NET_ADMIN" "CAP_NET_RAW"])

## Security

The service runs with:
- Minimal capabilities (NET_ADMIN and NET_RAW only)
- DynamicUser for isolation
- Comprehensive systemd sandboxing
- No privileged mode

## Development

Build the package:
```bash
nix build
```

Run the package:
```bash
nix run
```

## Integration with Prometheus

Add to your Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'ethtool'
    static_configs:
      - targets: ['localhost:9417']
``` 