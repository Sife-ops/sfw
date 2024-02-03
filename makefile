all: cw ww

cw:
	go build -o ./bin/cw ./cmd/cubiomes-worker/main.go

ww:
	go build -o ./bin/ww ./cmd/worldgen-worker/main.go
