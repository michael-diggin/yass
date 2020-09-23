proto:
	protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative api/yass.proto

build:
	docker build -t yass-server:0.1 .

run:
	docker run -d -p 8080:8080 yass-server:0.1


dev-start: build run

test:
	go test ./... --race --cover

hello:
	@echo "hello"