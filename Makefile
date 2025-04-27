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