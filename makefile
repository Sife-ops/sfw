.PHONY: all clean pkl cw ww mon

all: clean pkl cw ww mon

clean:
	rm -rf ./gen/config

pkl:
	pkl-gen-go ./pkl/config.pkl 

# todo add dependencies

cw:
	GOOS=linux GOARCH=amd64 go build -o ./bin/cw ./cmd/cubiomes_retvrn/main.go

ww:
	GOOS=linux GOARCH=amd64 go build -o ./bin/ww ./cmd/worldgen-worker/main.go

mon:
	GOOS=linux GOARCH=amd64 go build -o ./bin/mon ./cmd/monitor/main.go

# web:
# 	GOOS=linux GOARCH=amd64 go build -o ./bin/web ./cmd/web/main.go
