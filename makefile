all: cw ww log

cw:
	GOOS=linux GOARCH=amd64 go build -o ./bin/cw ./cmd/cubiomes-worker/main.go

ww:
	GOOS=linux GOARCH=amd64 go build -o ./bin/ww ./cmd/worldgen-worker/main.go

log:
	GOOS=linux GOARCH=amd64 go build -o ./bin/loggerino ./cmd/loggerino/main.go
