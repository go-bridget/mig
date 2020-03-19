package migrate

// List lists database migration filenames
func List() []string {
	result := make([]string, len(migrations))
	i := 0
	for k, _ := range migrations {
		result[i] = k
		i++
	}
	return result
}
