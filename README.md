# seed finding workers

## example hosts
```
[self]
localhost ansible_connection=local

[sfw_manager]
10.0.0.1 ansible_host=45.76.3.157

[sfw_new]
[sfw]
10.0.0.1 ansible_host=45.76.3.157
10.0.0.2 ansible_host=144.202.1.131

[sfw_managed]
10.0.0.2 ansible_host=144.202.1.131
```

## on local
```
ansible-playbook ./ansible/all.yml -e "wg_ip=<> pg_pass=<> pg_user=<> pg_db=<>"
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