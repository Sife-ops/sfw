# seed finding workers

## example `ansible/inventory.yml`
```yml
---
self:
  hosts:
    localhost:
      ansible_connection: local
      wg:
        role: client
        ip: 10.0.0.33
        peer_ip: 162.xxx.145.251
      sshcopy:
        ignore: true
sfw_manager:
  hosts:
    "35.243.214.251":
      wg:
        role: server
        ip: 10.0.0.10
        port: '51871'
      sshcopy:
        role: server
      nfs: 
        role: server
        host: 10.0.0.10
        allowed: 10.0.0.0/24
        path: /var/nfs/sfw
        mount: /tmp/sfw_nfs
sfw_node:
  hosts:
    "45.63.1.103":
      wg:
        role: client
        ip: 10.0.0.20
```

extract vultr hosts
```javascript
// todo update for inventory.yml
s = ""; document.querySelectorAll("[data-original-title='Copy IP Address']").forEach((x, i) => s = `${s}${x.innerText} wgip=10.0.0.${(i * 10) + 20}\n`); console.log(s)
```

## example `pkl/amends.pkl`
```pkl
amends "./config.pkl"

wgip = "my_hostname"

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

worldgen {
    ravine_proximity = 4
}
```

## todo
- pg_dump https://www.postgresql.org/docs/current/app-pgdump.html