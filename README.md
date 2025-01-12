<h1 align="center">
    Cupid
</h1>
<p align="center">
  Cupid is a Discord bot that allows users to send flowers and anonymous messages to other users.
</p>
<p align="center">
  Cupid also maintains a leaderboard of the most popular users on the server.
</p>

<p align="center">
  <img src="https://github.com/math280h/cupid/actions/workflows/pr.yaml/badge.svg" />
</p>

## Getting Started

Create a new bot on the [Discord Developer Portal](https://discord.com/developers/applications) and invite it to your server.

### Running the docker container

**NOTE:** In the .env file, it's important none of the values is sourrounded by quotes.
  If you are using qoutes docker will escape the values as \"value\" and the bot will not work.
  (*We will make a fix for this in the future*)

```bash
docker compose up
```

## Configuration

Cupid can be configured either through parameters passed to the bot or through environment variables.

You can find all env variables in the [.env.example](.env.example) file.

You can find all command line arguments [flags.go](internal/shared/flags.go)

## Development

### Prisma

```bash
go run github.com/steebchen/prisma-client-go db push
```
