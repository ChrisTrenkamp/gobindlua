# Generate struct bindings for GopherLua

`gobindlua` generates [GopherLua](https://github.com/yuin/gopher-lua) bindings for your structs.

`gobindlua` is designed to be used with `go:generate`.  For example:

```go
// Replace *version* with a gobindlua version.
//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua@*version* -s SomeStruct
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

* if the -s and -p parameter is unspecified, it should take the line/col set from go:generate (if set) to determine if it should generate a struct or package functions.
* gobindlua should be able to exclude fields and methods.
* See if it's possible to auto-generate documentation from the Go documentation on the struct, the struct fields, functions, and methods so it can be used with Lua LSP's (possibly with https://github.com/LuaLS/lua-language-server ?)
