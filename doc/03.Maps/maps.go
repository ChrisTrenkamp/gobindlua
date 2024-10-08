package maps

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type User struct {
	Name  string
	Age   int
	Email string
}

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type UserDatabase struct {
	Users map[int]User
}

//gobindlua:constructor
func NewUserDatabase() UserDatabase {
	return UserDatabase{
		Users: make(map[int]User),
	}
}

//gobindlua:constructor
func NewUserDatabaseFrom(users map[int]User) UserDatabase {
	return UserDatabase{
		Users: users,
	}
}

//gobindlua:constructor
func NewUser(name string, age int, email string) User {
	return User{
		Name:  name,
		Age:   age,
		Email: email,
	}
}
