package lib

import (
	"log"
	"math"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var Db *sqlx.DB

func init() {
	if _, e := os.Stat("./db.sqlite"); e != nil {
		log.Fatalf("%v", e)
	}
	db, e := sqlx.Open("sqlite3", "./db.sqlite")
	if e != nil {
		log.Fatalf("%v", e)
	}
	Db = db
}

type GodSeed struct {
	Id   int     `db:"id"`
	Seed *string `db:"seed"`

	RavineProximity *int `db:"ravine_proximity"`
	RavineChunks    *int `db:"ravine_chunks"`
	IronShipwrecks  *int `db:"iron_shipwrecks"`
	AvgBastionAir   *int `db:"abg_bastion_air"`

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

	Timestamp string `db:"timestamp"`
}

type Coords struct {
	X int
	Z int
}

func (g *GodSeed) RavineArea(block_offset int) (int, int, int, int) {
	// todo sus calcs
	return *g.ShipwreckX - block_offset,
		*g.ShipwreckZ - block_offset,
		*g.ShipwreckX + block_offset + 15,
		*g.ShipwreckZ + block_offset + 15
}

func (g *GodSeed) ShipwreckArea() (int, int, int, int) {
	return *g.ShipwreckX - 16,
		*g.ShipwreckZ - 16,
		*g.ShipwreckX + 31,
		*g.ShipwreckZ + 31
}

func (g *GodSeed) NetherChunksToBastion() (netherChunks2Load []Coords) {
	bz, bx := *g.BastionZ+8, *g.BastionX+8
	// log.Printf("info bastion chunk center coords %d,%d", bx, bz)
	s := float64(bz) / float64(bx)
	// log.Printf("info bastion slope %f", s)
	bxa := math.Abs(float64(bx))

	for i := 1; i < int(bxa); i++ {
		x := i
		if bx < 0 {
			x = i * -1
		}

		a, b := int(math.Floor(float64(x)/16)), int(math.Floor(float64(x)*s/16))
		hasChunk := false
		for _, v := range netherChunks2Load {
			if v.X == a && v.Z == b {
				hasChunk = true
			}
		}
		if hasChunk == false {
			netherChunks2Load = append(netherChunks2Load, Coords{a, b})
		}
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