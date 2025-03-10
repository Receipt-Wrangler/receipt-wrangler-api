package models

func BuildGroupMap() map[GroupRole]int {
	groupMap := make(map[GroupRole]int)
	groupMap[VIEWER] = 0
	groupMap[EDITOR] = 1
	groupMap[OWNER] = 2
	return groupMap
}

func BuildUserRoleMap() map[UserRole]int {
	userRoleMap := make(map[UserRole]int)
	userRoleMap[USER] = 0
	userRoleMap[ADMIN] = 1
	return userRoleMap
}

func HasRole(role UserRole, roleToCheck UserRole) bool {
	userRoleMap := BuildUserRoleMap()
	return userRoleMap[role] <= userRoleMap[roleToCheck]
}
