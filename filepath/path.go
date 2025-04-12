package filepath

import (
	"log"
	"path/filepath"
	"strings"
)

func Route(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.TrimPrefix(path, "./")
	path = path[:strings.LastIndex(path, ".")]
	path = "/" + path
	log.Println("filepath.Route", path)
	return path
}

func Format(path string) string {
	if filepath.Separator == '/' {
		// Linux
		log.Println("filepath.Format", path)
		return path
	}

	// Windows
	path = strings.ReplaceAll(path, "/", string(filepath.Separator))
	// path = strings.TrimPrefix(path, fmt.Sprintf(".%s", string(filepath.Separator)))
	log.Println("filepath.Format", path)
	return path
}
