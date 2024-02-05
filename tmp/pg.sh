DATA="experiment02"

docker run  --rm \
        --name sfw_db \
        -e POSTGRES_PASSWORD=todo \
        -e POSTGRES_USER=todo \
        -e POSTGRES_DB=todo \
        -e PGDATA=/var/lib/postgresql/data/pgdata \
        -p 5432:5432 \
        -v $(pwd)/${DATA}:/var/lib/postgresql/data \
        postgres