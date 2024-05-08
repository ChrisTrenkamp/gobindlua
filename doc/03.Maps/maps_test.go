package maps

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ChrisTrenkamp/gobindlua"
	lua "github.com/yuin/gopher-lua"
)

const script = `
local user1 = user.new("Mike Smith", 42, "mike.smith@example.com")
local user2 = user.new("Ryan Kennedy", 23, "rkennedy04021@nyu.com")
local user3 = user.new("Robert Rose", 70, "rrose00011@aol.com")
local db = user_database.new_from(
	{
		[10]=user1,
		[11]=user2,
		[12]=user3,
	}
)
user_db.users = db.users
`

func ExampleUserDatabase() {
	L := lua.NewState()
	defer L.Close()

	User{}.RegisterLuaType(L)
	UserDatabase{}.RegisterLuaType(L)
	gobindlua.RegisterLuaMap(L)

	user_db := UserDatabase{}
	L.SetGlobal("user_db", gobindlua.NewUserData(&user_db, L))

	if err := L.DoString(script); err != nil {
		log.Fatal(err)
	}

	jsonBytes, err := json.MarshalIndent(user_db.Users, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(jsonBytes))

	// Output:
	//{
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
