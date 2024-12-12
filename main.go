package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/kmtym1998/gh-prowl/cmd"
)

//go:embed notify/assets/*.mp3
var embedded embed.FS

func main() {
	defer func() {
		if r := recover(); r != nil {
			color.Red(fmt.Sprint(r))
			os.Exit(1)
		}
	}()

	f, err := embedded.Open("notify/assets/chime.mp3")
	if err != nil {
		panic("failed to open sound file:" + err.Error())
	}
	defer f.Close()

	ec, err := cmd.NewExecutionContext(f)
	if err != nil {
		panic("failed to initialize cli:" + err.Error())
	}

	if err := cmd.NewRootCmd(ec).Execute(); err != nil {
		panic(err)
	}
}
