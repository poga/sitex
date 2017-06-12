# SiteX

[![Go Report Card](https://goreportcard.com/badge/github.com/poga/sitex)](https://goreportcard.com/report/github.com/poga/sitex)

A static web server with support to Netlify's [redirect rules](https://www.netlify.com/docs/redirects/).

`go get github.com/poga/sitex`

## Usage

Given the following directory tree:

```
$ tree .
.
├── _redirects
├── test-2.json
└── test.json

0 directories, 3 files
```

and the `_redirect` file:

```
# redirect / to test.json
/ /test.json 200

# 301 redirect
/foo /test.json

# query params
/bar id=:id /test-:id.json
```

Start the server at the root of the directory:

```
$ sitex
```

Then

* `/` will render `/test.json`
* `/foo` will redirect to `/test.json`
* `/bar?id=2` will render `/test-2.json`

## Redirects

SiteX is built from scratch to mimic Netlify's `redirect` behavior. For documents, see [offical document](https://www.netlify.com/docs/redirects/).

## Contribute

Feel free to open an issue if you find difference between SiteX and Netlify.

## License

The MIT License

