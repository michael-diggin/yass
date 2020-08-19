build:
	GOOS=linux go build -o grpc_server github.com/michael-diggin/yass/backend/cmd
	docker build -t mdiggin/yass-server:0.1 .
	rm grpc_server

run:
	docker run -p 8080:8080 mdiggin/yass-server:0.1

test:
	go test ./...