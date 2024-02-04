# seed finding workers

## example hosts
```
[self]
localhost ansible_connection=local

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

## go
```
ansible-playbook ./playbook.yml
```
