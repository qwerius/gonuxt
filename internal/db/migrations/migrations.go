package migrations

type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

var Migrations = []Migration{
	Migration001Users,
	Migration002AddUpdatedAt,
	Migration003UpdateUsersUpdatedAt,
	Migration004FillUpdatedAt,
	Migration005CreateRole,
	Migration006SeedRoles,
	Migration007CreateUserRoles,
	Migration008AssignAdminToSuweri,
	Migration009ReassignAdminToSuweri,
	Migration010Profiles,
	Migration011UserProfiles,
}
