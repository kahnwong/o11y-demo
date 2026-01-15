start:
	go run .
start-lgtm:
	docker compose -f compose-otel.yaml up -d
start-beyla:
	sudo beyla --config beyla-config.yaml
test:
	hurl hurl/foo.hurl
