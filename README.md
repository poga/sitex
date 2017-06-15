# SiteX

[![Go Report Card](https://goreportcard.com/badge/github.com/poga/sitex)](https://goreportcard.com/report/github.com/poga/sitex)
[![codecov](https://codecov.io/gh/poga/sitex/branch/master/graph/badge.svg)](https://codecov.io/gh/poga/sitex)
[![Build Status](https://travis-ci.org/poga/sitex.svg?branch=master)](https://travis-ci.org/poga/sitex)

A static web server with support to Netlify's [redirect and rewrite rules](https://www.netlify.com/docs/redirects/), [custom headers, and basic auth](https://www.netlify.com/docs/headers-and-basic-auth/).

`go get github.com/poga/sitex`

## Usage

Run `sitex` in your site directory. For example:

```
$ git clone git@github.com:poga/sitex.git
$ cd sitex/example
$ sitex
```
Then you got a web server which:

* `http://localhost:8080/` will render `/test.json`
* `http://localhost:8080/test.json` will render the file
* `http://localhost:8080/foo` will redirect to `/test.json`
* `http://localhost:8080/bar?id=2` will render `/test-2.json`

## Redirects & Headers

You can define redirects by adding `_redirects` file to the root of your directory.

```
# redirect / to test.json
/ /test.json 200

# 301 redirect
/foo /test.json

# redirect when query params matches
/bar id=:id /test-:id.json
```

If you need to return custom headers or http basic authentication. add `_headers` file to the root of your directory.

```
## A path:
/templates/index.html
  # Headers for that path:
  X-Frame-Options: DENY
  X-XSS-Protection: 1; mode=block
/templates/index2.html
  X-Frame-Options: SAMEORIGIN

# match splat
/*
  X-Frame-Options: DENY
  X-XSS-Protection: 1; mode=block

# basic auth
/something/*
  Basic-Auth: someuser:somepassword anotheruser:anotherpassword

# match placeholder
/foo/:bar
  X-Frame-Options: DENY
```

SiteX is built from scratch to mimic Netlify's features. For detailed documents, see Netlify's [redirect document](https://www.netlify.com/docs/redirects/) and [header document](https://www.netlify.com/docs/headers-and-basic-auth/).

## Contribute

Feel free to open an issue if you find difference between SiteX and Netlify.

## License

The MIT License

