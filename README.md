# seed finding workers

## example hosts
vultr extract hosts
```javascript
s = ""; document.querySelectorAll("[data-original-title='Copy IP Address']").forEach(x => s = s + x.innerText + " sfwip=10.0.0.\n"); console.log(s)
```

```
[self]
localhost ansible_connection=local

[sfw_manager]
149.28.33.186 sfwip=10.0.0.10

# hosts in this tag will be rebooted
[sfw_new]
[sfw]
149.28.33.186 sfwip=10.0.0.10
149.28.46.78 sfwip=10.0.0.20

[sfw_node]
149.28.46.78 sfwip=10.0.0.20
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
./bin/cw -db_host=10.0.0.10:5432
./bin/ww -db_host=10.0.0.10:5432
```

## todo
- pg_dump https://www.postgresql.org/docs/current/app-pgdump.html