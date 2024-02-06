DATA="experiment01"

docker run  --rm \
        --name sfw_db \
        -e POSTGRES_PASSWORD=seed \
        -e POSTGRES_USER=seed \
        -e POSTGRES_DB=seed \
        -e PGDATA=/var/lib/postgresql/data/pgdata \
        -p 5432:5432 \
        -v $(pwd)/${DATA}:/var/lib/postgresql/data \
        postgres
