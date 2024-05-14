# Generate bindings for GopherLua

`gobindlua` generates [GopherLua](https://github.com/yuin/gopher-lua) bindings for structs and package functions.

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

It will also generate [LuaLS](https://github.com/LuaLS/lua-language-server) definitions under `lua_SomeStruct_definitions.go`:

```lua
---@meta some_struct

local some_struct = {}

---@return some_struct_fields
function some_struct.new() end

---@class some_struct_fields
---@field public my_string string[]
local my_struct = {}

---@return string
function some_struct_fields:join() end

return some_struct
```

## Tutorials

See [the docs](doc) for instructions on how to use `gobindlua`.

## Hacking gobindlua

When making changes to `gobindlua`, you can build and test it by running:

```
go generate ./... && go test ./...
```

## TODO for interfaces

Interface implementations have to be declared excplicitly with the -im parameter.  This can get tedious to maintain.  These -im parameters should be removed in favor of the following:

* `gobindlua` should read a `gobindlua-conf.json` file at the root of the project.  This file defines the Go modules that have `gobindlua` bindings.  `gobindlua` should load each of these packages and gather all of their interface declarations.  
* When generating a struct or function, and a field/param/return type is an interface, and that interface is within the list of modules and the interface has a `go:generate gobindlua` directive, make the Lua definition that interface type.  Otherwise, the Lua definition is the `any` type.
* When generating a struct, check if it implements any of the interfaces in the list of Go Modules.  If it implements that type, and the interface has a `go:generate gobindlua` directive, have the `@class` Lua definition for that struct extend that interface type.
