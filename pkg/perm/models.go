package perm

type Actor struct {
	ID        string
	Namespace string
}

type Permission struct {
	Action          string
	ResourcePattern string
}

type Role struct {
	Name string
}

type Action struct {
	Name string
}
