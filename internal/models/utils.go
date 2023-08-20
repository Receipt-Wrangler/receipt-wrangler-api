package models

func BuildGroupMap() map[GroupRole]int {
	groupMap := make(map[GroupRole]int)
	groupMap[VIEWER] = 0
	groupMap[EDITOR] = 1
	groupMap[OWNER] = 2
	return groupMap
}

func HasRole(role UserRole, roleToCheck UserRole) bool {
	return role == roleToCheck
}
