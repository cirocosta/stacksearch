# stacksearch

discover callstacks in which functions of interest are called.


## usage

1. collect a profile

```console
$ curl localhost:1337/debug/pprof/heap > ./heap.1.pprof
```


2. search for your function of interest

```console
$ stacksearch -p './heap.1.pprof' time.After
time.NewTimer
time.After
github.com/concourse/dex/server.(*Server).newHealthChecker.func1

time.NewTimer
time.After
github.com/concourse/dex/server.(*Server).startKeyRotation.func1

time.NewTimer
time.After
github.com/concourse/concourse/atc/scheduler.(*Runner).Run
github.com/tedsuo/ifrit.(*process).run
```

