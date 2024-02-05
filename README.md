# seed finding workers

## example hosts
```
[self]
localhost ansible_connection=local

[sfw_]

[sfw]
207.246.85.7
64.176.223.124
140.82.12.104

[sfw_manager]
207.246.85.7

[sfw_managed]
64.176.223.124
140.82.12.104
```

## on local
```
ansible-playbook ./playbook.yml
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
./bin/ww -db_host=104.207.132.120:5432 -db_name=todo000 -db_pass=todo000 -db_user=todo000 -inst=sfw0
```

## todo
- pg_dump https://www.postgresql.org/docs/current/app-pgdump.html