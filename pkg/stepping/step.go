package stepping

type Step struct {
	Name        string
	Description string
	Func        func() error
	Optional    bool
}
