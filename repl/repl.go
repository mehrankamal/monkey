package repl

import (
	"bufio"
	"fmt"
	"github.com/mehrankamal/monkey/compiler"
	"github.com/mehrankamal/monkey/lexer"
	"github.com/mehrankamal/monkey/object"
	"github.com/mehrankamal/monkey/parser"
	"github.com/mehrankamal/monkey/vm"
	"io"
	"strings"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	constants := make([]object.Object, 0)
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		c := compiler.NewWithState(symbolTable, constants)
		err := c.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}

		constStrings := []string{}

		for idx, constant := range c.Bytecode().Constants {
			constStrings = append(constStrings, fmt.Sprintf("%d: %s", idx, constant.Inspect()))
		}

		fmt.Fprintf(out, "Generated Opcodes:\n\t%s\nConstants:\n\t%s\nResults: ", strings.Join(strings.Split(c.Bytecode().Instructions.String(), "\n"), "\n\t"), strings.Join(constStrings, "\n\t"))

		code := c.Bytecode()
		constants = code.Constants

		machine := vm.NewWithGlobalsStore(code, globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}

		lastPopped := machine.LastPoppedStackElem()
		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")
	}
}

const MonkeyFace = `       .-"-.            .-"-.            .-"-.           .-"-.
     _/_-.-_\_        _/.-.-.\_        _/.-.-.\_       _/.-.-.\_
    / __} {__ \      /|( o o )|\      ( ( o o ) )     ( ( o o ) )
   / //  "  \\ \    | //  "  \\ |      |/  "  \|       |/  "  \|
  / / \'---'/ \ \  / / \'---'/ \ \      \'/^\'/         \ .-. /
  \ \_/'"""'\_/ /  \ \_/'"""'\_/ /      /'\ /'\         /'"""''\
   \           /    \           /      /  /|\  \       /       \

 -={ see no evil }={ hear no evil }={ speak no evil }={ have no fun }=-
`

func printParserErrors(out io.Writer, errors []string) {
	_, _ = io.WriteString(out, MonkeyFace)
	_, _ = io.WriteString(out, "Whoops! We ran into some monkey business here!\n")
	_, _ = io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		_, _ = io.WriteString(out, "\t"+msg+"\n")
	}
}
