# ListenStats

A listener-counting reverse proxy

**VERY EARLY**

## Background

ListenStats was built for us to handle counting listeners on our Icecast streams. We originally did it using [Mr. Freeze](https://github.com/UniversityRadioYork/mr_freeze), which used Icecast's listener authentication feature, and while it works, there were quite a number of issues we wanted to avoid:

* Lack of support for X-Forwarded-For
* Being tied to Icecast
* It's A Bit Of A Hack:tm:

ListenStats is an experimental replacement for this part of Mr. Freeze. It proxies Icecast, passing through all requests to it while recording them.

In theory, this should work with non-Icecast hosts - you could in theory use ListenStats as a rather overkill hit counter.

## Installation

Requires Go. Clone the Git repo, run `go get` and `go build`.

Copy `config.toml.example` to `config.toml` and customise it to your liking.

Then, just run `./listenstats`!

## About

ListenStats was developed for [University Radio York](https://github.com/UniversityRadioYork) by [Marks Polakovs](https://github.com/markspolakovs). Licensed under the MIT license.
