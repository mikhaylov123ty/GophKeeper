FROM golang:1.24 AS gobuild

WORKDIR /app

COPY . .

RUN go mod download

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

RUN go build -o ./certManager ./cmd/certManager/cert.go

RUN go build -ldflags "-X main.buildVersion=v1 -X "main.buildDate=$(date +"%Y/%m/%d")"" -o ./server ./cmd/server/main.go



FROM alpine

WORKDIR /app

COPY --from=gobuild /app/server ./GophServer
COPY --from=gobuild /app/cmd/server/config.json ./config.json
COPY --from=gobuild /app/./migrations ./migrations/

COPY --from=gobuild /app/certManager ./certManager

EXPOSE 4443

ENTRYPOINT ["/app/GophServer", "-config=./config.json"]