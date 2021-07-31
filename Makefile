genproto:
	protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative proto/yass.proto

test:
	go test ./... --race --cover
