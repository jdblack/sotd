Song of the day 


Usage
======
Create ~/.sotd.ini with the following contents:

```
[database]
type = "sqlite"
path = "file::memory:?cache=shared"

[slack]
botToken=xoxb-#############-#############-########################
appToken=xapp-#-###########-#############-################################################################
```


Databases
============

Sotd supports both SQLite and MySQL which may be chosen by setting type under
[Database] to "sqlite" or "mysql" 

Sqlite
--------

The sqlite connector requires onoe argument under the database section, the
database path

path - the path of the sqlite database.  A temporary database can be created in
memory by setting a path of "file::memory:?cache=shared"

Mysql
------

The mysql connector requires 5 options;   host, port, user, pass and db.  


