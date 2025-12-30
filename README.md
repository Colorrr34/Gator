# RSS Aggregator project - Gator

## Description

Gator is an CLI RSS aggregator tool which allows users to store RSS feeds in a PostgresQL database and save the posts in the feeds.

## Set up

This repo is using Go version 1.25.5 and PostgresQL database. Please install go and Postgres to run this program.<br>

pressly/goose package is used for migration in this repo
All schemas are in `sql/schema` directory.<br>

You can use go install command to install this CLI tool

```
go install github.com/colorrr34/gator
```

You need to create a .gatorconfig.json file at your home directory that will be used for database queries. <br>
The Database URL and current user will be stored in the config as following:

```
{
    "db_url": "DATABASE URL HERE",
    "current_user_name": "USERNAME HERE"
}
```

## Commands

Here are the commands available in Gator.

`register <username>` - register a user <br>

`login <username>` - login <br>

`reset` - remove all users <br>

`users` - show all users <br>

`addfeed <feed name> <url>` - add rss feed as the current user <br>

`feeds` - show all feeds for the current user <br>

`follow <url>` - follow feeds added by other users <br>

`following` - show all following feeds <br>

`unfollow <url>` - unfollow feed <br>

`agg <time_between_req>` - save all posts in your with a set time <br>

`browse <limit>` - show top posts in your saved posts, default limit is 2
