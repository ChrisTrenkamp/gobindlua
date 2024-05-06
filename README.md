# Generate struct bindings for GopherLua

`gobindlua` generates [GopherLua](https://github.com/yuin/gopher-lua) bindings for your structs.

`gobindlua` is designed to be used with `go:generate`.  For example:

```
//go:generate gobindlua -s SomeStruct
type SomeStruct struct {
    // ...
}
```

... this will generate a file called `lua_SomeStruct.go`.

## Installation

Make sure the absolute path to your `$GOPATH/bin` directory is in your `$PATH` (or wherever your Go binaries are installed).  `go:generate` will not work with relative paths.  e.g.:

```
export GOPATH="${HOME}/go"
export PATH="${GOPATH}/bin:${PATH}"
```

#### From https://pkg.go.dev/

```
go install github.com/ChrisTrenkamp/gobindlua/gobindlua@latest
```

#### From source

```
git clone https://github.com/ChrisTrenkamp/gobindlua
go build -o $GOPATH/bin/gobindlua gobindlua/gobindlua.go
```

## Tutorials

See [the docs](doc) for instructions on how to use `gobindlua`.

## Hacking gobindlua

When making changes to `gobindlua`, you can build and test it by running:

```
go build -o $GOPATH/bin/gobindlua gobindlua/gobindlua.go && go generate ./... && go test ./...
```

## TODO

* Add support for maps.
* Gather user types that are used in the struct, add them as dependencies, and auto-register them in the `RegisterLuaType` method.
* gobindlua should be able to forgo generating a struct, and only generate bindings for functions.
* gobindlua should be able to exclude fields and methods.
