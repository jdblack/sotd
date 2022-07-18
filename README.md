# Song of the day bot

SOTD Slack bot that provides daily (or more) links to a daily song of the day.
The bot provides a separate playlist for channel that it is.   Channel
participants are then widely encouraged to contribute songs to the channel
playlist.

The playlist for each channel is randomized, allowing users to add many of
their favorite songs at one time without hogging the front of the playlist
queue.

## Usage

Interation with the bot primarily happens via private message. You can send
the following commands to the bot to control song of the day for your channel


### General commands
-------------------

- `hi` or `hello` 
Check if the bot ia answering
- `channels` 
List the channels that the Bot is in

### Song Management commands

- `add #channel URL An optional user explanation for why they contributed this song`
add ads a new song to the playlist for the given channel. The song will also be used as
backfill for channels that have empty playlists
- `add! #channel URL Song description` - Add a song to the playlist even if the
  song has previously been added to sotd
- `song rm URL` [TODO] Completely remove a song from SOTDbot.  Intended for those NSFW moments

### Playlist Commands

- `playlist list` List out all of the channel playlists
- `playlist show CHANNEL_NAME` Show the playlist 
- `playlist delete CHANNEL_NAME` [TODO] tell the bot to delete a channel playlist and leave the channel. 
- `playlist cron CHANNEL_NAME  new crontab [TODO] Set a new schedule for the specified playlist. See Cron below for more information



## Crontab

Each channel can be scheduled independently of one another. The pattern is a
stanadrd cron format (https://en.wikipedia.org/wiki/Cron) with numeric fields
for Minute, Hour, Day of Month, Month and Day of week


` 00 22 * * 1-5 ` - Play a song at 22:00 UTC (3PM PDT) Mon-Fri  to celebrate the deployment window
` 00 16 * * 1 ` -  Play a monday morning song every Monday for the SRE Livesite channel 

The main intent is for each channel to set the best time for SOTD for their own
channel playlist, but the functiality is there if you want to set your cron up
for certain months or days of the month too.  


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
   bot --> controller : FromBot channel
   controller --> bot : ToBot channel
   jukebox --> controller : Playset channel
```


## Dedication

This project is dedicated to my close friends Belmin, Brandon, Brian, Drew,
Jaysen, Jeff, Jeremy and Jhurani.  To quote Martin Gore,  you made "All the
things I detest, I will almost like"  Thanks for making the last half decade of
my life such an adventure.  Call me when you need me. =)



