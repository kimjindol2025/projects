package main

import (
	"fmt"
	"fv2-lang/internal/typechecker"
)

func main() {
	checker := typechecker.New()
	
	// 빌트인 함수 확인
	println_sym := checker.GlobalScope.Lookup("println")
	print_sym := checker.GlobalScope.Lookup("print")
	
	fmt.Printf("println: %v\n", println_sym)
	fmt.Printf("print: %v\n", print_sym)
	
	if println_sym != nil {
		fmt.Printf("println type: %T - %s\n", println_sym.Type, println_sym.Type.TypeString())
	}
	
	if print_sym != nil {
		fmt.Printf("print type: %T - %s\n", print_sym.Type, print_sym.Type.TypeString())
	}
}
