slack-oldbot
=======

A bot for [Slack](https://slack.com) written in Go (golang) that politely reports when a link has been used before in the channel.

Usage
-----

* Build the code with `go build`

* Start the bot with `./slack-oldbot` on an internet-accessible server. (Check the output of `./slack-oldbot -h` for configuration options)

* Configure an [Outgoing Webhook](https://my.slack.com/services/new/outgoing-webhook) in your Slack and point it to the place where your bot is running. For example: `http://example.com:8000/`

* The bot will listen to incoming requests, extract urls, add them to the list, and if they've been used before, it will respond in the channel.

Tips
----

* Export your team's data from https://my.slack.com/services/export, and use that to seed the bot. See the `-importDir` and `-importChan` options.
* Keep your bot scoped to one channel.
