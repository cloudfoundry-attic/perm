package perm

type Actor struct {
	ID        string
	Namespace string
	Groups    []Group
}

type Group struct {
	ID string
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
