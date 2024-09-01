package lion

//go:generate go run github.com/ChrisTrenkamp/gobindlua/gobindlua
type Lion struct{}

//gobindlua:constructor
func NewLion() Lion {
	return Lion{}
}

//gobindlua:function
func (c Lion) Sound() string {
	return "rawr"
}
