# kneissbot
👑 moderator management tool for Twitch using DPoS

Kneissbot is a moderator management tool for Twitch streamers. It promotes a democratic, community-driven system through staking and social interaction.
* leverages delegated proof of stake and a cryptographically secure ledger
* parses IRC messages using regular expressions
* places trust in viewers to decide what is best for the community by electing delegates
* utilizes moving averages to gauge the trend of the chat
* estimates required moderators based on a heuristic

## Usage
`go run kneissbot.go`

The bot will run a temporary, local server for you to authenticate with Twitch and retrieve an authorization token.

Viewers will need to !register with the bot.  
Viewers who wish to be moderator need to become a !delegate.  
Viewers can also !vote for delegates.  

## What this project does
Kneissbot aids in having moderators available at all times, having enough moderators to handle demand, and having the best interest of the stream.
* moderators are elected by the community
* recognizes community members who contribute to the stream
* scales to n moderators based on demand
* updates every t seconds distributing rewards, checking delegates, and adding / removing moderators

## Why this project exists
Streams are community based. Communities can grow to exceedingly large numbers. It can be hard to have everyone act in accordance to the community guidelines.
* [Too Many Moderators](https://www.reddit.com/r/Twitch/comments/37g08k/too_many_moderators/)
* difficulty moderating several messages per second
* moderators are not available during all hours of a stream
