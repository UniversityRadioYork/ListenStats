# ListenStats

A listener-counting reverse proxy

**VERY EARLY**

## Background

ListenStats was built for us to handle counting listeners on our Icecast streams. We are currently doing it using [Mr. Freeze](https://github.com/UniversityRadioYork/mr_freeze), which uses Icecast's listener authentication feature, and while it works, there were quite a number of issues we would like to avoid:

* Lack of support for X-Forwarded-For
* Being tied to Icecast
* It's A Bit Of A Hack:tm:

ListenStats is an experimental replacement for this part of Mr. Freeze. It proxies Icecast, passing through all requests to it while recording them.

In theory, this should work with non-Icecast hosts - you could in theory use ListenStats as a rather overkill hit counter.

## Installation

## From Binaries

Download a [binary](https://github.com/UniversityRadioYork/ListenStats/releases) of the latest release for your OS of choice. If you want to use a bleeding-edge build, you can download them from [GitHub Actions](https://github.com/UniversityRadioYork/ListenStats/actions?query=workflow%3AGo) (click on the latest build and look under Artifacts), but be aware that they may be buggy!

Download [config.toml.example](https://raw.githubusercontent.com/UniversityRadioYork/ListenStats/master/config.toml.example) and save it as `config.toml`, and edit it to your needs.

Then, run `./listenstats-{version}-{os}`. You can pass in a path to an alternate config file as the first parameter, e.g. `./listenstats /path/to/config.toml`.

### From Source

Requires Go. Clone the Git repo, run `go get` and `go build`.

Copy `config.toml.example` to `config.toml` and customise it to your liking.

Then, just run `./listenstats`! You can pass in a path to an alternate config file as the first parameter, e.g. `./listenstats /path/to/config.toml`.

## About

ListenStats was developed for [University Radio York](https://github.com/UniversityRadioYork) by [Marks Polakovs](https://github.com/markspolakovs). Licensed under the MIT license.
