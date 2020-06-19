package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CompileDir(dir string, outputDir string) error {
	if len(outputDir) == 0 {
		outputDir = dir
	}
	var sources []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".jack" {
			sources = append(sources, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, file := range sources {
		if err := CompileFile(file, outputDir); err != nil {
			return err
		}
	}
	return nil
}

func CompileFile(file string, dir string) error {
	f, _ := os.Open(file)
	defer f.Close()
	tokenizer := &tokenizer{}
	tokenizer.Tokenize(f)
	analysizer := &analysizer{}
	output, err := analysizer.LexialAnalysis(tokenizer.tokens)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	jc, err := parseClass(output)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	cw := NewVmCompiler(jc)
	base := filepath.Base(file)
	fo, err := os.Create(fmt.Sprintf("%s/%s.vm", dir, strings.TrimSuffix(base, filepath.Ext(file))))
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer fo.Close()
	cw.compileAndWrite(fo)
	fmt.Println(fmt.Sprintf("%s compiled to %s", file, dir))
	return nil
}
