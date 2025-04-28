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

build:
	go build -gcflags="-N -l" -o ./cmd/client/app ./cmd/client/main.go

runApp:
	./cmd/client//app -config cmd/client/config.json

run:
	go run cmd/client/main.go -config ./cmd/client/config.json