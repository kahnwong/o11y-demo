start:
	go run .
start-lgtm:
	docker compose -f compose-otel.yaml up -d
test:
	hurl hurl/get.hurl
	hurl hurl/post.hurl
