# A tcp checker

[![Build Status](https://travis-ci.com/cloverstd/tcp-checker.svg?branch=master)](https://travis-ci.com/cloverstd/tcp-checker)

# Install

```bash
go install github.com/cloverstd/tcp-checker
```

# Usage

```golang
import (
    "github.com/cloverstd/tcp-checker"
)

func main() {
    checker, err := tcpchecker.New(Option{
        DefaultDown: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    if checker.Down("hui.lu:80") {
        log.Println("hui.lu:80 is down")
    }
}
```
