package tfbicep

import (
	"os"
	"path/filepath"
)

// FileInfo holds info about a discovered IaC file
type FileInfo struct {
	Path string
	Type string // "tf", "bicep", "tfstate"
}

// ScanIaCFiles scans a directory recursively for .tf, .bicep, and .tfstate files
func ScanIaCFiles(root string) ([]FileInfo, error) {
	var files []FileInfo
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		switch filepath.Ext(info.Name()) {
		case ".tf":
			files = append(files, FileInfo{Path: path, Type: "tf"})
		case ".bicep":
			files = append(files, FileInfo{Path: path, Type: "bicep"})
		case ".tfstate":
			files = append(files, FileInfo{Path: path, Type: "tfstate"})
		}
		return nil
	})
	return files, err
}
