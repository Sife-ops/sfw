.PHONE: clean genpkl all

all: cw ww log web mon

clean:
	rm -rf ./gen/config

genpkl:
	pkl-gen-go ./pkl/config.pkl

# todo add dependencies

cw:
	GOOS=linux GOARCH=amd64 go build -o ./bin/cw ./cmd/cubiomes-worker/main.go

ww:
	GOOS=linux GOARCH=amd64 go build -o ./bin/ww ./cmd/worldgen-worker/main.go

log:
	GOOS=linux GOARCH=amd64 go build -o ./bin/loggerino ./cmd/loggerino/main.go

web: main.go
	GOOS=linux GOARCH=amd64 go build -o ./bin/web ./cmd/web/main.go

mon:
	GOOS=linux GOARCH=amd64 go build -o ./bin/mon ./cmd/monitor/main.go
