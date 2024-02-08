# seed finding workers

## example hosts
vultr extract hosts
```javascript
s = ""; document.querySelectorAll("[data-original-title='Copy IP Address']").forEach((x, i) => s = `${s}${x.innerText} os=arch sfwip=10.0.0.${(i * 10) + 20}\n`); console.log(s)
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

[sfw_node_new]
[sfw_node]
149.28.46.78 sfwip=10.0.0.20
```

## example usage
```
./bin/cw -db_host=<> -db_pass=<>
./bin/ww
```

## todo
- pg_dump https://www.postgresql.org/docs/current/app-pgdump.html