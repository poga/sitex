# SiteX

[![Go Report Card](https://goreportcard.com/badge/github.com/poga/sitex)](https://goreportcard.com/report/github.com/poga/sitex)
[![codecov](https://codecov.io/gh/poga/sitex/branch/master/graph/badge.svg)](https://codecov.io/gh/poga/sitex)
[![Build Status](https://travis-ci.org/poga/sitex.svg?branch=master)](https://travis-ci.org/poga/sitex)

A static web server with support to Netlify's [redirect and rewrite rules](https://www.netlify.com/docs/redirects/), [custom headers, and basic auth](https://www.netlify.com/docs/headers-and-basic-auth/).

`go get -u github.com/poga/sitex`

## Usage

Run `sitex` in your site directory. For example:

```
$ git clone git@github.com:poga/sitex.git
$ cd sitex
$ sitex -dir example
Serving example/ at :8080
```
Now you got a web server which:

* `http://localhost:8080/` will render `/test.json`
* `http://localhost:8080/test.json` will render the file
* `http://localhost:8080/foo` will redirect to `/test.json`
* `http://localhost:8080/bar?id=2` will render `/test-2.json`

#### CLI options

* dir: the directory you want to server. **Default: current working directory**.
* port: port to listen. **Default: 8080**.

## Rules

For example, see `_headers` and `_redirects` file in the `example` folder.

SiteX is built from scratch to mimic Netlify's features. For detailed documents, see Netlify's [redirect document](https://www.netlify.com/docs/redirects/) and [header document](https://www.netlify.com/docs/headers-and-basic-auth/).

## Contribute

Feel free to open an issue if you find difference between SiteX and Netlify.

## License

The MIT License

