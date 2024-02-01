all: sch cw wg

db:
	sqlite3 db_.sqlite < ./sql/table.sql

sch: db
	go build -o ./bin/scheduler ./cmd/scheduler/main.go

cw: db
	go build -o ./bin/cubiomes-worker ./cmd/cubiomes-worker/main.go

wg: db
	go build -o ./bin/worldgen-worker ./cmd/worldgen-worker/main.go
