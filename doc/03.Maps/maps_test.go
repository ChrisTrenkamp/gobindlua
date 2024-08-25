package maps

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

const script = `
local user = require "user"
local user_database = require "user_database"
local gbl_map = require "gbl_map"

local user1 = user.new("Mike Smith", 42, "mike.smith@example.com")
local user2 = user.new("Ryan Kennedy", 23, "rkennedy04021@nyu.com")
local user3 = user.new("Robert Rose", 70, "rrose00011@aol.com")

--[[ Just like slices, you can use tables to construct Go maps. ]]
local db = user_database.new_from(
	{
		[10]=user1,
		[11]=user2,
	}
)

print("db size: " .. tostring(#db.users))
db.users[12]=user3

print("Directly indexing the Go map:")
print("db.users[10]: " .. db.users[10].name)
print("db.users[11]: " .. db.users[11].name)
print("db.users[12]: " .. db.users[12].name)

user_db.users = db.users

--[[ And you can convert maps back to tables. ]]
local db_table = gbl_map.to_table(db.users)

print("Indexing the Lua table:")
for k,v in pairs(db_table) do
	print("db_table[" .. tostring(k) .. "] = " .. v.name)
end
`

func Example() {
	L := lua.NewState()
	defer L.Close()

	gobindlua.Register(L, &User{}, &UserDatabase{})

	user_db := UserDatabase{}
	L.SetGlobal("user_db", gobindlua.NewUserData(&user_db, L))

	if err := L.DoString(script); err != nil {
		log.Fatal(err)
	}

	jsonBytes, err := json.MarshalIndent(user_db.Users, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Map result in Go:", string(jsonBytes))

	// Output:
	//db size: 2
	//Directly indexing the Go map:
	//db.users[10]: Mike Smith
	//db.users[11]: Ryan Kennedy
	//db.users[12]: Robert Rose
	//Indexing the Lua table:
	//db_table[10] = Mike Smith
	//db_table[11] = Ryan Kennedy
	//db_table[12] = Robert Rose
	//Map result in Go: {
	//	"10": {
	//		"Name": "Mike Smith",
	//		"Age": 42,
	//		"Email": "mike.smith@example.com"
	//	},
	//	"11": {
	//		"Name": "Ryan Kennedy",
	//		"Age": 23,
	//		"Email": "rkennedy04021@nyu.com"
	//	},
	//	"12": {
	//		"Name": "Robert Rose",
	//		"Age": 70,
	//		"Email": "rrose00011@aol.com"
	//	}
	//}
}
