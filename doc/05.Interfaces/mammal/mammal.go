package mammal

import "github.com/ChrisTrenkamp/gobindlua"

// You can generate Lua definitions for interfaces by attaching a //go:generate directive
// to an interface.

// In order to pass around interfaces, they must implement gobindlua.LuaUserData.
// The interface implementation doesn't necessarily need to be generated with
// gobindlua, but its metadata table must be globally available.  Otherwise, it
// will not work.

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type Mammal interface {
	Sound() string
	gobindlua.LuaUserData
}
