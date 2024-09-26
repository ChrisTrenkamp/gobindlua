package maps

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

const script = `
local User = require "User"
local UserDatabase = require "UserDatabase"
local GblMap = require "GblMap"

local user1 = User.NewUser("Mike Smith", 42, "mike.smith@example.com")
local user2 = User.NewUser("Ryan Kennedy", 23, "rkennedy04021@nyu.com")
local user3 = User.NewUser("Robert Rose", 70, "rrose00011@aol.com")

--[[ Just like slices, you can use tables to construct Go maps. ]]
local db = UserDatabase.NewUserDatabaseFrom(
	{
		[10]=user1,
		[11]=user2,
	}
)

print("db size: " .. tostring(#db.Users))
db.Users[12]=user3

print("Directly indexing the Go map:")
print("db.users[10]: " .. db.Users[10].Name)
print("db.users[11]: " .. db.Users[11].Name)
print("db.users[12]: " .. db.Users[12].Name)

user_db.Users = db.Users

--[[ And you can convert maps back to tables. ]]
local db_table = GblMap.ToTable(db.Users)

print("Indexing the Lua table:")
for k,v in pairs(db_table) do
	print("db_table[" .. tostring(k) .. "] = " .. v.Name)
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
