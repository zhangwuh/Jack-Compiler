# Jack-Compiler

This is a Jack compiler implemented in go. It consists of two parts:

##Syntax Analysis
1. tokenizer: tokenizer.go
2. Lexial analyzer: analysizer.go

##Code generation
1.Implement symbol table: vm_variables.go
2.Build analysis results to structured object to make it more smoothly when transfer tokenized retuls to vm code
3.Compile object to vm code: vm_code_generator.go

Usage: 
