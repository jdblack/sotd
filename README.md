# Song of the day bot

This project is a Song of the Day Slack Bot that provides daily link from a 
list of songs to one or more channels.   The bot provides a separate 
playlist for each channel.   Channels that run out of their private list of
songs will backfill from songs that have been previously played in other
channels.

The playlist for each channel is randomized, allowing users to add many of
their favorite songs at one time without hogging the front of the playlist
queue. Prolific users are able to swamp out other users,  so it's 
recommended that you widely encourage many people to participate!

## Usage

Interation with the bot primarily happens via private message. You can send
the following commands to the bot to control song of the day for your channel


### Bot commands
-------------------

This bot only listens for private messaages.  


*add CHANNEL URL [Song description]* Add a song to a play with optional
description

*delete URL* Delete song matching URL

*playlists* List all running playlists

*load CHANNEL URL* Import a json playlist from an URL.  Imports will be
credited to the importer

*stop CHANNEL* Tell SOTD to remove a playlist for a channel. The songs will be
saved for backfill, but the playlist will be gone, gone, gone

*show CHANNEL* Show the playlist for a given channel

*hello* Say hello


## Crontab

Each channel can be scheduled independently of one another. The pattern is a
stanadrd cron format (https://en.wikipedia.org/wiki/Cron) with numeric fields
for Minute, Hour, Day of Month, Month and Day of week (numeric, sunday being 0)


` 00 22 * * 1-5 ` - Play a song at 22:00 UTC (3PM PDT) Mon-Fri  to celebrate the deployment window
` 00 16 * * 1 ` -  Play a monday morning song every Monday for the SRE Livesite channel 

The main intent is for each channel to set the best time for SOTD for their own
channel playlist, but the functiality is there if you want to set your cron up
for certain months or days of the month too.  

## Bulk Importing Songs

Songs can be imported over HTTP via a json formatted list. The list should have
the following structure:

```
[
  {
    "URL": "https://www.youtube.com/watch?v=hQKfAdhXBpA"
    "Description": "Kids playing drums in new orleans"
  },
  {
    "URL": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "Description": "Would I be me without this one?"
  },
]
```



## Installation

Create ~/.sotd.ini with the following contents:

```
[database]
type = "sqlite"
path = "file::memory:?cache=shared"

[slack]
botToken=xoxb-#############-#############-########################
appToken=xapp-#-###########-#############-################################################################
```


## Databases

Sotd supports both SQLite and MySQL which may be chosen by setting type under
[Database] to "sqlite" or "mysql" 

Sqlite
--------

The sqlite connector requires  single argument under the database section, the
path of the database. A temporary database can be created in memory by setting
a path of "file::memory:?cache=shared"

```
[database]
type = "sqlite"
# path = "file::memory:?cache=shared" # temporary in-memory db for testing
path = "/var/lib/sotdbot.db"
```

Mysql
------

The mysql connector requires 5 options;   host, port, user, pass and db.  

```
[database]
type = "mysql"
user = "username"
host = "mysql-server"
pass = "password"
db   = "database name"
```


Design
--------

```mermaid
graph TD;
   slack{slack api}<-- event handler -->bot
   bot-- FromBot<br />channel --> controller
   controller-- ToBot<br />channel -->bot
   controller-- funcs  -->jukebox
   controller-- funcs  -->bot
   jukebox --> cron{Cron Scheduler}
   cron{Cron Scheduler}-- Playset ch -->controller
```

## Dedication

This project is dedicated to my close friends Belmin, Brandon, Brian, Drew,
Jaysen, Jeff, Jeremy and Jhurani.  To quote Martin Gore,  you made "All the
things I detest, I will almost like"  Thanks for making the last half decade of
my life such an adventure.  Call me when you need me. =)


