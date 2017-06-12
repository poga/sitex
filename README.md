# SiteX

[![Go Report Card](https://goreportcard.com/badge/github.com/poga/sitex)](https://goreportcard.com/report/github.com/poga/sitex)

A static web server with support to Netlify's [redirect rules](https://www.netlify.com/docs/redirects/).

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

## Redirects

SiteX is built from scratch to mimic Netlify's `redirect` behavior. For documents, see [offical document](https://www.netlify.com/docs/redirects/).

## Contribute

Feel free to open an issue if you find difference between SiteX and Netlify.

## License

The MIT License

