# QueryMap

QueryMap is a small Go library that allows you to parse a Query String URL into a convenient 
structure (nested-Map / slices), and then convert that data into Go structures as needed.

## Features

- Recognizes nested parameters of `key[a][b]` format, automatically creating nested structures.
- Supports repeated keys (for example: `key=value1&key=value2`), combining their values into slices.
- Allows converting parsing results into structures via `mapstructure`.
- Automatically detects sequences of numeric keys (`0`, `1`, `2`, etc.), converting them into slices.
- Can handle all types of Go query-parameters (`string`, `[]string`, `map[string]any`, etc.).
- Allows convenient conversion of parsing results into structures (using [mapstructure](https://github.com/mitchellh/mapstructure) package).

## Installation

```bash
go get github.com/KoNekoD/go-querymap
```

## Example

```go
package main

import (
	"fmt"
	"github.com/KoNekoD/go-querymap"
	"log"
)

type MyQueryParams struct {
	User string `json:"user"`
	Info struct {
		Age  int    `json:"age"`
		City string `json:"city"`
	} `json:"info"`
	Tags []string `json:"tags"`
}

func main() {
	rawURL := "http://example.com?user=John&info[age]=30&info[city]=New+York&tags[]=go&tags[]=programming"

	result, err := querymap.FromURLStringToStruct[MyQueryParams](rawURL)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", result)
	// Output:
	// &{User:John Info:{Age:30 City:New York} Tags:[go programming]}
}
```

## Documentation

See comments in code and function:
- `FromURL`
- `FromURLToStruct`
- `FromURLStringToStruct`
- `ToStruct`

They all help you work with Query parameters in different ways.
