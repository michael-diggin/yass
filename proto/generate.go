//go:generate protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative yass.proto
//go:generate protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative watchtower.proto

//go:generate mockgen -destination=../common/client/mocks/mock_grpc_client.go -package=mocks . StorageClient
//go:generate mockgen -destination=../server/mocks/mock_watchtower_client.go -package=mocks . WatchTowerClient

package proto
