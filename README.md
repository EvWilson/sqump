# sqump
[![Tests](https://github.com/EvWilson/sqump/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/EvWilson/sqump/actions/workflows/test.yml)

`sqump` is a utility for managing small, composable scripts that perform various requests and operations on the results of these requests.
It is born out of a dissatisfaction with Postman, and after recent policy changes, an outright aversion.

## Quickstart
```
$ go install github.com/EvWilson/sqump@latest
$ sqump init
$ sqump exec Squmpfile.json NewReq
hello, world!
```

The above sequence should get you spun up and executing your first script! (Assuming you have Go 1.21+ installed.)
Check out `sqump help` to find out what's possible, or use `sqump webview` for a view to help explore what `sqump` has to offer.

## Documentation
Check out the [docs](docs) directory for more information about the Lua modules provided.

Also an important note! This area also holds the [walkthrough](docs/walkthrough.md) that is highly encouraged for all first-time users of `sqump`.

## FAQs
- Sqump?

[Sqump](https://youtu.be/MS1jJzoMUjI?si=PPH_hONo0wEKNAmx&t=414)
