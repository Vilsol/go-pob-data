package main

import (
	"os"

	_ "github.com/andybalholm/brotli"
)

//go:generate msgp -v -file ./raw

func main() {
	if len(os.Args) < 2 {
		println("please provide path to the game directory")
		os.Exit(1)
		return
	}

	if len(os.Args) < 3 {
		println("please provide passive tree version")
		os.Exit(1)
		return
	}

	if len(os.Args) < 4 {
		println("please provide game version")
		os.Exit(1)
		return
	}

	gamePath := os.Args[1]
	treeVersion := os.Args[2]
	gameVersion := os.Args[3]

	if err := extractRawData(gamePath, gameVersion); err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	if err := downloadTreeData(treeVersion, gameVersion); err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	if err := extractTranslations(gamePath, gameVersion); err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}
}
