package utils

import "receipt-wrangler/api/internal/models"

func BuildGroupMap() map[models.GroupRole]int {
	groupMap := make(map[models.GroupRole]int)
	groupMap[models.VIEWER] = 0
	groupMap[models.EDITOR] = 1
	groupMap[models.OWNER] = 2
	return groupMap
}
