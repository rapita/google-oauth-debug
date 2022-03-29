# Google OAuth debug

This repository contains simple web server for receiving callback by google oauth.

## Getting started

Make code:
```shell
git clone git@github.com:rapita/google-oauth-debug.git && cd google-oauth-debug
```

Install modules:
```shell
go mod tidy
```

Copy configs:
```shell
cp config.dist.yaml config.yaml
```
Change configs in `config.yaml`

Run server:
```shell
go run main.go
```