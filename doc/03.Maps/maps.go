package maps

//go:generate gobindlua -s User
type User struct {
	Name  string
	Age   int
	Email string
}

//go:generate gobindlua -s UserDatabase
type UserDatabase struct {
	Users map[int]User
}

func NewUserDatabase() UserDatabase {
	return UserDatabase{
		Users: make(map[int]User),
	}
}

func NewUserDatabaseFrom(users map[int]User) UserDatabase {
	return UserDatabase{
		Users: users,
	}
}

func NewUser(name string, age int, email string) User {
	return User{
		Name:  name,
		Age:   age,
		Email: email,
	}
}
