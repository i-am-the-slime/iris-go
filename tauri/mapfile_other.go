//go:build !unix

package tauri

import "os"

func mapFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}
