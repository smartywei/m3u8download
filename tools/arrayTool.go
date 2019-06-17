package tools

func InArray(needValue string, arr []string) bool {

	for _, v := range arr{
		if v == needValue {
			return true
		}
	}
	return false
}

