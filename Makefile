#common vars
DSN:="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

#generate proto files
generateProto:
	protoc --go_out=. --go_opt=paths=source_relative \
	  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	  internal/proto/handlers.proto

generateCert:
	go run ./cmd/certManager/cert.go


goimports:
	goimports -local github.com/mikhaylov123ty/GophKeeper -w ./internal/..

buildClient:
	go build -gcflags="-N -l" -o ./cmd/client/app ./cmd/client/main.go

runBuildClient: buildClient
	./cmd/client/app -config cmd/client/config.json

run Client:
	go run cmd/client/main.go -config ./cmd/client/config.json

buildServer:
	go build -gcflags="-N -l" -o ./cmd/server/app ./cmd/server/main.go

runBuildServer: buildServer
	./cmd/server/app -config cmd/server/config.json

run Server:
	go run cmd/server/main.go -config ./cmd/server/config.json