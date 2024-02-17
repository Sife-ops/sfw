package lib

import (
	"fmt"
	"log"
	"math"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var Db *sqlx.DB

func init() {
	if err := loadConfig(); err != nil {
		log.Printf("%v", err)
	}

	var connStr string
	if len(Cfg.Postgres.GetPassword()) > 0 {
		connStr = fmt.Sprintf(
			"postgres://%s:%s@%s/%s?sslmode=disable",
			Cfg.Postgres.GetUsername(),
			Cfg.Postgres.GetPassword(),
			Cfg.Postgres.GetHost(),
			Cfg.Postgres.GetDatabase(),
		)
	} else {
		connStr = fmt.Sprintf(
			"postgres://%s@%s/%s?sslmode=disable",
			Cfg.Postgres.GetUsername(),
			Cfg.Postgres.GetHost(),
			Cfg.Postgres.GetDatabase(),
		)
	}

	// todo pgx
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("error connect %v", err)
	}
	Db = db
}

type GodSeed struct {
	Id   int     `db:"id"`
	Seed *string `db:"seed"`

	RavineProximity     *int `db:"ravine_proximity"`
	RavineChunks        *int `db:"ravine_chunks"` // todo deprecated
	ExposedRavineBlocks *int `db:"exposed_ravine_blocks"`
	IronShipwrecks      *int `db:"iron_shipwrecks"`
	AvgBastionAir       *int `db:"avg_bastion_air"`

	Played *int `db:"played"`
	Rating *int `db:"rating"`

	SpawnX     *int `db:"spawn_x"`
	SpawnZ     *int `db:"spawn_z"`
	BastionX   *int `db:"bastion_x"`
	BastionZ   *int `db:"bastion_z"`
	ShipwreckX *int `db:"shipwreck_x"`
	ShipwreckZ *int `db:"shipwreck_z"`
	FortressX  *int `db:"fortress_x"`
	FortressZ  *int `db:"fortress_z"`

	FinishedCubiomes *int `db:"finished_cubiomes"`
	FinishedWorldgen *int `db:"finished_worldgen"`

	TstzCreated string `db:"tstz_created"`
}

type Coords struct {
	X int
	Z int
}

func (g *GodSeed) RavineAreaX1() int {
	return *g.ShipwreckX - Cfg.Worldgen.RavineProximity*16
}

func (g *GodSeed) RavineAreaZ1() int {
	return *g.ShipwreckZ - Cfg.Worldgen.RavineProximity*16
}

func (g *GodSeed) RavineAreaX2() int {
	return *g.ShipwreckX + Cfg.Worldgen.RavineProximity*16 + 15
}

func (g *GodSeed) RavineAreaZ2() int {
	return *g.ShipwreckZ + Cfg.Worldgen.RavineProximity*16 + 15
}

func (g *GodSeed) ShipwreckAreaX1() int {
	return *g.ShipwreckX - 16
}

func (g *GodSeed) ShipwreckAreaZ1() int {
	return *g.ShipwreckZ - 16
}

func (g *GodSeed) ShipwreckAreaX2() int {
	return *g.ShipwreckX + 31
}

func (g *GodSeed) ShipwreckAreaZ2() int {
	return *g.ShipwreckZ + 31
}

func (g *GodSeed) NetherChunksToBastion() (netherChunks2Load []Coords) {
	bz, bx := *g.BastionZ+8, *g.BastionX+8
	s := float64(bz) / float64(bx)
	bxa := math.Abs(float64(bx))

Outer:
	for i := 1; i < int(bxa); i++ {
		x := i
		if bx < 0 {
			x = i * -1
		}

		a, b := int(math.Floor(float64(x)/16)), int(math.Floor(float64(x)*s/16))
		for _, v := range netherChunks2Load {
			if v.X == a && v.Z == b {
				continue Outer
			}
		}
		
		netherChunks2Load = append(netherChunks2Load, Coords{a, b})
	}

	return
}

func MustInt(i int, err error) int {
	if err != nil {
		panic(err)
	}
	return i
}

func MustIntRef(i int, err error) *int {
	if err != nil {
		panic(err)
	}
	return &i
}

func ToStringRef(i string) *string {
	return &i
}

func ToIntRef(i int) *int {
	return &i
}

func MustString(i string, err error) string {
	if err != nil {
		panic(err)
	}
	return i
}

func ToSector(i int) int {
	if i < 0 {
		return 32 + i
	}
	return i
}
