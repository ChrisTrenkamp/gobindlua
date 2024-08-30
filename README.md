# Generate bindings for GopherLua

`gobindlua` generates [GopherLua](https://github.com/yuin/gopher-lua) bindings and [LuaLS](https://github.com/LuaLS/lua-language-server) definitions.  It can generate bindings and definitions for structs, interfaces, and package functions.

## Example

`gobindlua` is designed to be used with `go:generate`:

```go
// Replace *version* with a gobindlua version.

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua@*version*
type Joiner interface {
    Join() string
}

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua@*version*
type SomeStruct struct {
    SomeStrings []string
}

//gobindlua:constructor
func NewSomeStruct(strs []string) SomeStruct {
    return SomeStruct {
        SomeStrings: strs,
    }
}

//gobindlua:function
func (s SomeStruct) Join() string {
    return strings.Join(s.SomeStrings, ", ")
}
```

... this will generate a file called `lua_SomeStruct.go`.  The generated bindings will work seamlessly with Lua tables:

```lua
local some_struct = require "some_struct"
local my_struct = some_struct.new({"foo", "bar", "eggs", "ham"})
print(my_struct:join()) --[[ foo, bar, eggs, ham ]]
```

It will also generate [LuaLS](https://github.com/LuaLS/lua-language-server) definitions for the struct and interface:

```lua
---lua_SomeStruct_definitions.go
---@meta some_struct

local some_struct = {}

---@return some_struct_fields
function some_struct.new() end

---@class some_struct_fields : joiner
---@field public my_string string[]
local my_struct = {}

---@return string
function some_struct_fields:join() end

return some_struct
```

```lua
---lua_Joiner_definitions.go
---@meta joiner

---@class joiner
local joiner = {}

---@return string
function joiner.join() end

return joiner
```

## Enable interface discovery

If you want to generate interface definitions, create a `gobindlua-conf.txt` file in the root of your Go project (next to the `go.mod` file), with the list of Go modules that have generated `gobindlua` source files (including your own project).  Any interface field listed in the `gobindlua-conf.txt` and has a `//go:generate` directive will be picked up and generated as its own type in the Lua definitions.  Otherwise, it will be generated as an `any` type.

## Tutorials

See [the docs](doc) for instructions on how to use `gobindlua`.

## Hacking gobindlua

When making changes to `gobindlua`, you can build and test it by running:

```
go generate ./... && go test ./...
```

TODO:

* `godoc` comments for structs, fields, and methods should be attached to the lua definitions.
