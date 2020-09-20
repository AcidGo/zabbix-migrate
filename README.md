# zabbix-migrate
```
Options:
  -c string
    	select the type of check, support for hostgroup|host|item|trigger|valuemap|map|all
  -d uint
    	input params about day offset (default 1)
  -f string
    	set path of config file than ini format (default "zabbix_migrate.ini")
  -g string
    	input params about hostgroup
  -h	show for help
  -htable string
    	select the name of history table for sync
  -i int
    	input params about host begin id
  -ignore
    	ignore migrate errors
  -l uint
    	set log level number, 0 is panic ... 6 is trace (default 4)
  -m string
    	select the type of migrate, support for hostgroup|valuemap|template|host
  -o uint
    	input params about id offset (default 50)
  -s string
    	select the type of sync, support for trends|history
```
