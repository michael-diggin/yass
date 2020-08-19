build:
	docker build -t mdiggin/yass-server:0.1 .

run:
	docker run -p 8080:8080 mdiggin/yass-server:0.1

test:
	go test ./...