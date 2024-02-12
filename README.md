# seed finding workers

## example `ansible/hosts`
```
[self]
localhost ansible_connection=local

[sfw_manager]
149.28.33.186 sfwip=10.0.0.10

[sfw_node]
149.28.46.78 sfwip=10.0.0.20
```

extract vultr hosts
```javascript
s = ""; document.querySelectorAll("[data-original-title='Copy IP Address']").forEach((x, i) => s = `${s}${x.innerText} sfwip=10.0.0.${(i * 10) + 20}\n`); console.log(s)
```

## example `pkl/amends.pkl`
```pkl
amends "./config.pkl"

postgres: Postgres = new {
    host = "10.0.0.10:5432"
    database = "seed"
    username = "seed"
    password = "seed"
}

log: Log = new {
    host = "10.0.0.10:1337"
}

web: Web = new {
    host = "127.0.0.1:3000"
}
```

## example usage
```
./bin/cw -db_host=<> -db_pass=<>
./bin/ww
```

## todo
- pg_dump https://www.postgresql.org/docs/current/app-pgdump.html