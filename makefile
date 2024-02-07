all: cw ww log

cw:
	go build -o ./bin/cw ./cmd/cubiomes-worker/main.go

ww:
	go build -o ./bin/ww ./cmd/worldgen-worker/main.go

log:
	go build -o ./bin/loggerino ./cmd/loggerino/main.go
