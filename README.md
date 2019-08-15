# v0v.cc: a URL Shortener Service

This is a URL shortener service written in Go. It powers `v0v.cc`.

# API

To shorten a URL, send a POST request with it as the url field in the body. For example, to shorten `https://google.com`, run:
```
$ curl -d "url=google.com&auth=XXX" v0v.cc
```
where the output is `v0v.cc/BST`. `auth` is the authentication code needed. If the URL given does not have a scheme, then it is assumed to be https. Internationalized hostnames are supported.

URLs are redirected using 301 Permanent redirects.

# Build

Install the [bolt](https://github.com/boltdb/bolt) library using:
```
go get "github.com/boltdb/bolt"
```

Initalized your config like `config.go.example`. Run 
```
go build server.go config.go -o shortener
```

It probably requires root privileges.

# License

This project is released into the public domain. [Favicon](https://www.favicon.cc/?action=icon&file_id=686272) used under Creative Commons, no attribution license.
