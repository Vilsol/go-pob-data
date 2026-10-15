package extractor

import (
	"fmt"
	"strconv"
	"strings"
)

// Game is the client the data comes from.
type Game string

const (
	PoE1 Game = "poe1"
	PoE2 Game = "poe2"
)

func ParseGame(s string) (Game, error) {
	switch g := Game(s); g {
	case PoE1, PoE2:
		return g, nil
	}
	return "", fmt.Errorf("unknown game %q, want %q or %q", s, PoE1, PoE2)
}

// validFor is g's bit in dat-schema's validFor (1 PoE1, 2 PoE2, 3 both).
func (g Game) validFor() int {
	if g == PoE2 {
		return 2
	}
	return 1
}

// TableDir is the bundle directory that holds g's tables.
func (g Game) TableDir() string {
	if g == PoE2 {
		return "Data/Balance/"
	}
	return "Data/"
}

// PoE2 versions start at 0 and have always used the new hash and .datc64, so only PoE1 reads the minor.

func IsNewHash(game Game, gameVersion string) bool {
	if game == PoE2 {
		return true
	}
	n, _ := strconv.Atoi(strings.Split(gameVersion, ".")[1])
	return n >= 22
}

func GetFileExtension(game Game, gameVersion string) string {
	if game == PoE2 {
		return "datc64"
	}
	n, _ := strconv.Atoi(strings.Split(gameVersion, ".")[1])
	if n >= 25 {
		return "datc64"
	}
	return "dat64"
}
