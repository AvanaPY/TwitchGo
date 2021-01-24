# TwitchGO

#### A module for Go that allows simple usage of Twitch's Chatbot API. 
* Inspired by the Python library `Discord.py` which is a similar library but for Discord's API
* Simple to use
* Made solely to try out Golang
* Personal project

# Information

## Types

* Channel
    * Struct used to store information about a Twitch channel
    * Created automatically by the Client type

* Client
    * The Client
    * Runs the program and handles messages from Twitch
    * User may interact with the Client and JOIN Twitch Chatbot rooms during runtime
    * User may create commands at compile time or runtime
    * Automatically handles Twitch's `PING` request to keep the connection alive

* Command
    * A user-created command
    * Created by the user
    * Automatically invoked by the Client when it receives the appropriate message
    * When invoked it calls a function defined by the user.

* Context
    * Stores information about a message received from Twitch to the Client
        * Who sent it
        * Where was it sent
        * What was sent
        * What command it shall invoke
    * Created automatically by the Client type upon receiving a message from Twitch
