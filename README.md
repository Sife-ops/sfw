# seed finding workers

## example hosts
```
[self]
localhost ansible_connection=local

[sfw_manager]
66.135.19.43 sfwip=10.0.0.10

[sfw_new]
[sfw]
66.135.19.43 sfwip=10.0.0.10
66.135.22.113 sfwip=10.0.0.20

[sfw_managed]
66.135.22.113 sfwip=10.0.0.20
```

## on local
```
ansible-playbook ./ansible/all.yml -e "pg_pass=<> pg_user=<> pg_db=<>"
```

## on db
```
psql -h localhost -U <db_user> -d <db_name> -a -f ./sql/pg.sql
psql -U <db_user> -d <db_name> -a -f ./sql/pg.sql
```

## on workers
```
./build.sh
./bin/cw -db_host=<>:5432 -db_name=<> -db_pass=<> -db_user=<> -inst=sfw<>
./bin/ww -db_host=<>:5432 -db_name=<> -db_pass=<> -db_user=<> -inst=sfw<>
./bin/cw -db_host=127.0.0.1:5432 -db_name=todo000 -db_pass=todo000 -db_user=todo000 -inst=sfw0
./bin/ww -db_host=10.0.0.10:5432 -db_name=todo000 -db_pass=todo000 -db_user=todo000 -inst=sfw0
```

## todo
- pg_dump https://www.postgresql.org/docs/current/app-pgdump.html