package filter

//go:generate mockgen -destination types_mock.go -package filter . Blocker
type Blocker interface {
	Block(string)
	Unblock(string)
	List() []string
}
