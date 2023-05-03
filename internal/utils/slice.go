package utils

func Contains(slice []interface{}, target interface{}) bool {
	for _, value := range slice {
		if value == target {
			return true
		}
	}

	return false
}
