CREATE TABLE seed (
    id SERIAL PRIMARY KEY,
    seed TEXT,

    ravine_proximity INTEGER,
    ravine_chunks INTEGER,
    iron_shipwrecks INTEGER,
    avg_bastion_air INTEGER,
    -- avg_fortress_air INTEGER,
    played INTEGER,
    rating INTEGER,

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

    tstz_created TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
