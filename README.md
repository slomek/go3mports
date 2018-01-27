# go3mports

Pronunciation: _go-three-mports_

The idea behind this tool is to go through the Go files in the directory and determine if the imports are grouped correctly. The most common approach   is to have them separated into two categories:

```
import (
    // stdlib
    "fmt"
    "strings"

    // the rest
    "github/pkg/errors"
    "github/slomek/mappy"
)
```

I prefer to have more granularity, so that the _rest_ is sepated into 3rd party dependencies and something that I (or the organization that the repo belongs to) created:

```
import (
    // stdlib
    "fmt"
    "strings"

    // 3rd party
    "github/pkg/errors"

    // in-house deps
    "github/slomek/mappy"
)
```

## Installation

```
go get github.com/slomek/go3mports
```

## Usage

In order to check grouping in given directory and all subdirectories, run:

    $ go3mports

For example, running this command from the root directory of this repository prints out the information about errors in `./example/example.go`:

    $ go3mports
    example/example.go: invalid grouping
    example/performers/performers.go: multiple import groups of the same type
