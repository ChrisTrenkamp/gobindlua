# Generate struct bindings for GopherLua

`gobindlua` generates [GopherLua](https://github.com/yuin/gopher-lua) bindings for your structs.

`gobindlua` is designed to be used with `go:generate`.  For example:

```go
// Replace *version* with a gobindlua version.
//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua@*version*
type SomeStruct struct {
    SomeStrings []string
}

func NewSomeStruct(strs []string) SomeStruct {
    return SomeStruct {
        SomeStrings: strs,
    }
}

Func (s SomeStruct) Join() string {
    return strings.Join(s.SomeStrings, ", ")
}
```

... this will generate a file called `lua_SomeStruct.go`.  The generated bindings will work seamlessly with Lua tables:

```lua
local my_struct = some_struct.new({"foo", "bar", "eggs", "ham"})
print(my_struct:join()) --[[ foo, bar, eggs, ham ]]
```

## Tutorials

See [the docs](doc) for instructions on how to use `gobindlua`.

## Hacking gobindlua

When making changes to `gobindlua`, you can build and test it by running:

```
go generate ./... && go test ./...
```

## TODO

* Interfaces implementations have to be declared excplicitly with the -im parameter.  This can get tedious.  Investigate a way to automatically detect if a struct implements an interface and have the class extend them in the Lua definitions.  Or at the very least, make a utility program that verifies the declarations are correct, or are missing entries.
