module github.com/michael-diggin/yass

go 1.14

require (
	github.com/armon/go-metrics v0.3.9 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.5.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/hashicorp/go-msgpack v1.1.5 // indirect
	github.com/hashicorp/raft v1.1.1
	github.com/hashicorp/raft-boltdb v0.0.0-20210422161416-485fa74b0b01
	github.com/hashicorp/serf v0.9.5
	github.com/soheilhy/cmux v0.1.5
	github.com/stretchr/testify v1.7.0
	github.com/tysontate/gommap v0.0.0-20210506040252-ef38c88b18e1
	go.uber.org/zap v1.17.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9 // indirect
	google.golang.org/genproto v0.0.0-20200423170343-7949de9c1215 // indirect
	google.golang.org/grpc v1.27.1
	google.golang.org/protobuf v1.27.1
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

replace github.com/hasicorp/raft-boltdb => github.com/travisjeffery/raft-boltdb v1.0.0
