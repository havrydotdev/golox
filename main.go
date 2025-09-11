package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/havrydotdev/golox/expr"
	"github.com/havrydotdev/golox/parser"
	"github.com/havrydotdev/golox/scanner"
)

func main() {
	if len(os.Args) > 1 {
		fileName := os.Args[1]
		text, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}

		tokens, err := scanner.New(string(text)).Scan()
		if err != nil {
			fmt.Printf("Scanning failed: %s\n", err.Error())
		}

		exprs, errs := parser.New(tokens, expr.NewEval()).Parse()
		for _, err := range errs {
			fmt.Println(err.Error())
		}

		for _, expr := range exprs {
			err := expr.Eval()
			if err != nil {
				fmt.Printf("Eval error: %s\n", err.Error())
				continue
			}
		}
	} else {
		fmt.Println("Welcome to GoLox (version 0.0.1)!")
		for {
			fmt.Print("> ")
			reader := bufio.NewReader(os.Stdin)
			text, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				continue
			}

			tokens, err := scanner.New(text).Scan()
			if err != nil {
				fmt.Printf("Scanning failed: %s\n", err.Error())
			}

			exprs, errs := parser.New(tokens, expr.NewEval()).Parse()
			for _, err := range errs {
				fmt.Println(err.Error())
			}

			for _, expr := range exprs {
				err := expr.Eval()
				if err != nil {
					fmt.Printf("Eval error: %s\n", err.Error())
					continue
				}
			}
		}
	}
}
