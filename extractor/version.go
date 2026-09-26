package extractor

import (
	"strconv"
	"strings"
)

func IsNewHash(gameVersion string) bool {
	n, _ := strconv.Atoi(strings.Split(gameVersion, ".")[1])
	return n >= 22
}

func GetFileExtension(gameVersion string) string {
	n, _ := strconv.Atoi(strings.Split(gameVersion, ".")[1])
	if n >= 25 {
		return "datc64"
	}
	return "dat64"
}
