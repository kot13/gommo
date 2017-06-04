# MMO about rabbits

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