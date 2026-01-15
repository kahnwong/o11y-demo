# o11y-demo

## Pre-reqs

Install Beyla

```bash
curl -LO https://github.com/grafana/beyla/releases/download/v2.8.4/beyla-linux-amd64-v2.8.4.tar.gz
tar xzf beyla-linux-amd64-v2.8.4.tar.gz
sudo mv beyla /usr/local/bin/
```

## Usage

```bash
make start-lgtm  # http://localhost:3000 - admin:admin
make start-beyla
make start # http://localhost:8080
```

## Notes

- See `logs` under `loki`
- See `metrics` under `prometheus` and select metric name
- See `traces` under `tempo`
- See supported Go frameworks: <https://github.com/grafana/beyla>

## Refs

- <https://oneuptime.com/blog/post/2025-12-10-ebpf-with-opentelemetry-auto-instrumentation/view>
