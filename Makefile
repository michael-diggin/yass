build:
	docker build -t mdiggin/yass-server:0.1 .

run:
	docker run -p 8080:8080 mdiggin/yass-server:0.1

redis:
	docker run --name redis-image -p 6379:6379 -d redis

test:
	go test ./...