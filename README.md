Song of the day 

# Usage
==============

Interation with the bot primarily happens via private message. You can send
the following commands to the bot to control song of the day for your channel


-`hi` or `hello` Check if the bot ia answering
-`channels` List the channels that the Bot is in

- `add #channel URL Your description of a song`
add a song to sotd bot
-`song rm URL` [TODO] Completely remove a song from SOTDbot.  Intended for those NSFW moments

-`playlist list` List out all of the channel playlists
-`playlist show CHANNEL_NAME` Show the playlist 
-`playlist delete CHANNEL_NAME` [TODO] tell the bot to delete a channel playlist and leave the channel. 
-`playlist cron CHANNEL_NAME  new crontab [TODO] Set a new schedule for the specified playlist. See Cron below for more information



# Crontab

Each channel can be scheduled independently of one another. The pattern is a
stanadrd cron format (https://en.wikipedia.org/wiki/Cron) with numeric fields
for Minute, Hour, Day of Month, Month and Day of week


` 00 22 * * 1-5 ` - Play a song at 22:00 UTC (3PM PDT) Mon-Fri  to celebrate the deployment window
` 00 16 * * 1 ` -  Play a monday morning song every Monday for the SRE Livesite channel 

The main intent is for each channel to set the best time for SOTD for their own
channel playlist, but the functiality is there if you want to set your cron up
for certain months or days of the month too






# Installation
=============
Create ~/.sotd.ini with the following contents:

```
[database]
type = "sqlite"
path = "file::memory:?cache=shared"

[slack]
botToken=xoxb-#############-#############-########################
appToken=xapp-#-###########-#############-################################################################
```


# Databases
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


[database]
type = "mysql"
user = "username"
host = "mysql-server"
pass = "password"
db   = "database name"


