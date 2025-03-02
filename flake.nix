{
  description = "Prometheus ethtool exporter";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "prometheus-ethtool-exporter";
          version = "0.1.0";
          src = ./.;

          vendorHash = "sha256-5wf8OslCSUem7q1M43pqILqbYqo8mNnqwjywB1OS9PI=";

          meta = with pkgs.lib; {
            description = "Prometheus exporter for ethtool metrics";
            homepage = "https://github.com/minhuw/prometheus-ethtool-exporter";
            license = licenses.mit;
            maintainers = [ ];
            platforms = platforms.linux;
          };
        };
      }
    ) // {
      nixosModules.default = { config, lib, pkgs, ... }:
        with lib;
        let
          cfg = config.services.prometheus.exporters.ethtool;
        in
        {
          options.services.prometheus.exporters.ethtool = {
            enable = mkEnableOption (mdDoc "prometheus ethtool exporter");

            port = mkOption {
              type = types.port;
              default = 9417;
              description = mdDoc "Port to listen on";
            };

            interfaces = mkOption {
              type = types.listOf types.str;
              default = [];
              description = mdDoc "List of network interfaces to monitor";
            };

            openFirewall = mkOption {
              type = types.bool;
              default = false;
              description = mdDoc "Open port in firewall for incoming connections";
            };

            capabilities = mkOption {
              type = types.listOf types.str;
              default = [ "CAP_NET_ADMIN" "CAP_NET_RAW" ];
              description = mdDoc "Required Linux capabilities for the service";
            };
          };

          config = mkIf cfg.enable {
            systemd.services.prometheus-ethtool-exporter = {
              description = "Prometheus ethtool exporter";
              wantedBy = [ "multi-user.target" ];
              after = [ "network.target" ];

              serviceConfig = {
                ExecStart = "${self.packages.${pkgs.system}.default}/bin/prometheus-ethtool-exporter";
                DynamicUser = true;
                AmbientCapabilities = cfg.capabilities;
                CapabilityBoundingSet = cfg.capabilities;
                # Security hardening
                NoNewPrivileges = true;
                ProtectSystem = "strict";
                ProtectHome = true;
                PrivateTmp = true;
                PrivateDevices = true;
                ProtectClock = true;
                ProtectControlGroups = true;
                ProtectKernelLogs = true;
                ProtectKernelModules = true;
                ProtectKernelTunables = true;
                RestrictAddressFamilies = [ "AF_INET" "AF_INET6" "AF_NETLINK" ];
                RestrictNamespaces = true;
                RestrictRealtime = true;
                RestrictSUIDSGID = true;
                MemoryDenyWriteExecute = true;
                LockPersonality = true;
              };
            };

            networking.firewall.allowedTCPPorts = mkIf cfg.openFirewall [ cfg.port ];
          };
        };
    };
} 