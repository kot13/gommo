# MMO about rabbits

[ ![Codeship Status for kot13/gommo](https://app.codeship.com/projects/67bf45f0-2b34-0135-7e95-4afd89638027/status?branch=master)](https://app.codeship.com/projects/224028)

## Run
You need to install gvt before:
```
$ go get -u github.com/FiloSottile/gvt
```
Now:
```
$ gvt restore
$ go build
$ ./gommo
```

Type in browser:
```
localhost:8080
```

## Embedded build
You need to install go.rice before:
```
$ go get github.com/GeertJohan/go.rice
$ go get github.com/GeertJohan/go.rice/rice
```
Now:
```
$ rice embed-go
$ go build
$ ./gommo
```

Or:
```
$ rice embed-go && go build && ./gommo
```
