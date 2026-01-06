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
make start-lgtm  # http://localhost:8080 - admin:admin
make start-beyla
make start # http://localhost:3000
```

## Notes

- See `logs` under `loki`
- See `traces` under `tempo`
- See `metrics` under `prometheus` and select metric name

## Refs

- <https://oneuptime.com/blog/post/2025-12-10-ebpf-with-opentelemetry-auto-instrumentation/view>
