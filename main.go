package main

import (
	"fmt"
	"os"

	"github.com/zhangwuh/jack-compiler/compiler"
)

func main() {
	args := os.Args
	if len(args) < 2 || len(args) > 3 {
		fmt.Println("invalid params, usage go run main.go [source dir] [output dir](optional)")
		return
	}
	sourcePath := args[1]
	var outputPath string
	if len(args) == 3 {
		outputPath = args[2]
	}
	if err := compiler.CompileDir(sourcePath, outputPath); err != nil {
		fmt.Println("compile err:" + err.Error())
	}
	fmt.Println("compile done")
}
