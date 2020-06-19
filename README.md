# Jack-Compiler

This is a Jack compiler implemented in golang followed by the book of "The Elements of Computing Systems"(By Noam Nisan and Shimon Schocken (MIT Press)). It consists of two parts:

## Syntax Analysis
1. tokenizer: tokenizer.go
2. Lexial analyzer: analysizer.go

## Code generation
1.Implement symbol table: vm_variables.go
2.Build analysis results to structured object to make it more smoothly when transfer tokenized retuls to vm code
3.Compile object to vm code: vm_code_generator.go

## Usage: go run main.go [source path] [output path(optional)]

You can run the compiled vm files with the vm emulator published by https://www.nand2tetris.org/
