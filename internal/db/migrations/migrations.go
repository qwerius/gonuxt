package migrations

type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

var Migrations = []Migration{
	Migration001Users,
}
