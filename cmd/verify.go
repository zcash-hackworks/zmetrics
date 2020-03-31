package cmd

import "os"

func tryWritingFile(path string) (err error) {
	_, err = os.Stat(path)
	if err != nil {
		return err
	}
	return nil
}
