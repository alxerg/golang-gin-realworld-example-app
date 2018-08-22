# ![RealWorld Example App](logo.png)


[![Build Status](https://travis-ci.org/wangzitian0/golang-gin-starter-kit.svg?branch=master)](https://travis-ci.org/wangzitian0/golang-gin-starter-kit)
[![codecov](https://codecov.io/gh/wangzitian0/golang-gin-starter-kit/branch/master/graph/badge.svg)](https://codecov.io/gh/wangzitian0/golang-gin-starter-kit)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/wangzitian0/golang-gin-starter-kit/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/wangzitian0/golang-gin-starter-kit?status.svg)](https://godoc.org/github.com/wangzitian0/golang-gin-starter-kit)

> ### Golang/Gin/Slowpoke codebase containing real world examples (CRUD, auth, advanced patterns, etc) that adheres to the [RealWorld](https://github.com/gothinkster/realworld) spec and API.


This codebase was created to demonstrate a fully fledged fullstack application built with **Golang/Gin** including CRUD operations, authentication, routing, pagination, and more.


This app use [slowpoke](https://github.com/recoilme/slowpoke) as database. Package slowpoke implements a low-level key/value store in pure Go.

![slowpoke](http://tggram.com/media/recoilme/photos/file_488344.jpg)


# Read it


don't use this code in production
I just ported this project https://github.com/gothinkster/golang-gin-realworld-example-app
from sqlite to my database - but i don't do any code review

My app, as this app - absolutely ugly. It does tons of database requests on each web request
That's a terrible part of code what i don't fix, sorry for that.. I feel shame I just learned golang on this project.

I fixed it and write high load optimized code in another project, similar to real world - https://github.com/recoilme/tgram

# How it works
```
.
├── gorm.db
├── hello.go
├── common
│   ├── utils.go        //small tools function
│   └── database.go     //DB connect manager
└── users
    ├── models.go       //data models define & DB operation
    ├── serializers.go  //response computing & format
    ├── routers.go      //business logic & router binding
    ├── middlewares.go  //put the before & after logic of handle request
    └── validators.go   //form/json checker
```

# Getting started

## Install the Golang
https://golang.org/doc/install
## Environment Config
make sure your ~/.*shrc have those varible:
```
➜  echo $GOPATH
/Users/zitwang/test/
➜  echo $GOROOT
/usr/local/go/
➜  echo $PATH
...:/usr/local/go/bin:/Users/zitwang/test//bin:/usr/local/go//bin
```
## Install Govendor & Fresh
I used Govendor manage the package, and Fresh can help build without reload

https://github.com/kardianos/govendor

https://github.com/pilu/fresh
```
cd 
go get -u github.com/kardianos/govendor
go get -u github.com/pilu/fresh
```

## Start
```
➜  govendor sync
➜  govendor add +external
➜  fresh
```


## Testing
```
➜  go test -v ./... -cover
```

## Todo
- More elegance config
- Test coverage (common & users 100%, article 0%)
- ProtoBuf support
- Code structure optimize (I think some place can use interface)
- Continuous integration (done)
