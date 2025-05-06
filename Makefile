#var
client_name = "GophKeeper"
version = 1
date = ${shell date -u +%Y/%m/%d}

#generate proto files
generateProto:
	protoc --go_out=. --go_opt=paths=source_relative \
	  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	  internal/proto/handlers.proto

generateCert:
	go run ./cmd/certManager/cert.go

goimports:
	goimports -local github.com/mikhaylov123ty/GophKeeper -w ./internal/..


run Client:
	go run cmd/client/main.go -config ./cmd/client/config.json

buildServer:
	go build -o ./cmd/server/server ./cmd/server/main.go

runBuildServer: generateCert buildServer
	./cmd/server/server -config cmd/server/config.json

run Server:
	go run cmd/server/main.go -config ./cmd/server/config.json


#build clients
buildClientMacIntel:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.buildVersion=v${version} -X "main.buildDate=${date}"" -o ./cmd/client/${client_name}_darwin_amd64 ./cmd/client/main.go

buildClientMac:
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.buildVersion=v${version} -X "main.buildDate=${date}"" -o ./cmd/client/${client_name}_darwin_arm64 ./cmd/client/main.go

buildClientWin:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.buildVersion=v${version} -X "main.buildDate=${date}"" -o ./cmd/client/${client_name}_windows_amd64.exe ./cmd/client/main.go

buildClientLinuxAMD:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.buildVersion=v${version} -X "main.buildDate=${date}"" -o ./cmd/client/${client_name}_linux_amd64 ./cmd/client/main.go

buildClientLinux:
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.buildVersion=v${version} -X "main.buildDate=${date}"" -o ./cmd/client/${client_name}_linux_arm64 ./cmd/client/main.go

buildClients: buildClientMacIntel buildClientMac buildClientWin buildClientLinuxAMD buildClientLinux