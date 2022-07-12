# Humi PHP Structured Logger

## What

A structured logger for use in our PHP projects.

<figure>
    <img width="300" src="logs.jpg"
         alt="logs">
</figure>

## Why

1. To unify logging across our platforms.
1. To make logging better.

## Installation

```sh
go get github.com/Humi-HR/go-structured-logger
```

```go
import logger "github.com/Humi-HR/go-structured-logger"
```

## Usage

### Basic

The most basic form of using this logger is as follows:

```go
// create a config
cfg := Config{
  Writer:  os.Stdout,           // io.Writer
  Env:     "production",        // should probably be sourced from the environment
  Service: "some-humi-service", // name of your service
}

lgr := NewLogger(cfg) // create your logger instance
defer lgr.Flush()     // defer flushing

lgr.Info("We are logging!")                                                // regular log
lgr.Info("We are logging with context!").WithContext(`{"message": "Yay!"`) // log with context
```

### Request Middleware

If you are using this logger in an web API, you might consider using the built-in middleware.

The middleware configures a new instance of the logger per request and comes with a helper to retrieve the logger from the request.

```go
// create a config
cfg := Config{
Writer: os.Stdout, // io.Writer
Env: "production", // should probably be sourced from the environment
Service: "some-humi-service", // name of your service
}

mdl := logger.Middleware(cfg) // configure the middleware
router.Use(mdl) // register the middleware

lgr, err := logger.FromContext(r.Context()) // retrieve the logger from the request context
if err != nil {
// handle misconfiguration
}

lgr.Info("We are logging!") // regular log
lgr.Info("We are logging with context!").WithContext(`{"message": "Yay!"`) // log with context
```
