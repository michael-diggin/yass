proto:
	protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative proto/yass.proto

build:
	docker build -t yass-server:0.1 .

run:
	docker run -d -p ${PORT}:${PORT} yass-server:0.1 -p ${PORT}

dev-start: build run

test:
	go test ./... --race --cover
