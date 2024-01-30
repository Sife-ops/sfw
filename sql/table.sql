CREATE TABLE seed (
    id INTEGER PRIMARY KEY,
    seed TEXT NOT NULL UNIQUE,
    ravine_chunks INTEGER NOT NULL,
    iron_shipwrecks INTEGER NOT NULL,
    avg_bastion_air INTEGER NOT NULL,
    avg_fortress_air INTEGER NOT NULL,
    played INTEGER DEFAULT 0 NOT NULL,
    rating INTEGER,
    notes TEXT,

	spawn_x INTEGER,
	spawn_z INTEGER,
	bastion_x INTEGER,
	bastion_z INTEGER,
	shipwreck_x INTEGER,
	shipwreck_z INTEGER,
	fortress_x INTEGER,
	fortress_z INTEGER,

    finished_cubiomes INTEGER,
    finished_worldgen INTEGER,

    timestamp TEXT DEFAULT CURRENT_TIMESTAMP NOT NULL
);