package tools

import "os"

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateDir(path string) error {

	pathExist, err := PathExists(path)

	if err != nil{
		return err
	}

	if !pathExist {
		err := os.MkdirAll(path, 0755)

		if err != nil {
			return err
		}
	}
	return nil
}