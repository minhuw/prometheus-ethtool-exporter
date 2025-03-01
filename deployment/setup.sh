#!/bin/bash

# Get Docker host IP
HOST_IP=$(ip -4 addr show docker0 | grep -Po 'inet \K[\d.]+')
if [ -z "$HOST_IP" ]; then
    HOST_IP="172.17.0.1"  # Default fallback
fi

# Update prometheus.yml with correct host IP
sed -i "s/172.17.0.1:9417/$HOST_IP:9417/" prometheus/prometheus.yml

echo "Updated prometheus.yml with host IP: $HOST_IP"
echo "You can now run: docker compose up -d" 