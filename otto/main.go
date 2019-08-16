package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	// "github.com/dlclark/regexp2"
	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/underscore"
)

var flag_underscore *bool = flag.Bool("underscore", true, "Load underscore into the runtime environment")

func readSource(filename string) ([]byte, error) {
	if filename == "" || filename == "-" {
		return ioutil.ReadAll(os.Stdin)
	}
	return ioutil.ReadFile(filename)
}

func main() {
	// regularExpression2, err2 := regexp2.Compile("a(?!b)", 0x0100)
	// if err2 != nil {
	// 	fmt.Printf(err2.Error())
	// }
	// res, _ := regularExpression2.MatchString("ac")
	// if res {
	// 	fmt.Printf("match!!!\n")
	// } else {
	// 	fmt.Printf("no match\n")
	// }

	flag.Parse()

	if !*flag_underscore {
		underscore.Disable()
	}

	err := func() error {
		src, err := readSource(flag.Arg(0))
		if err != nil {
			return err
		}

		vm := otto.New()
		_, err = vm.Run(src)
		return err
	}()
	if err != nil {
		switch err := err.(type) {
		case *otto.Error:
			fmt.Print(err.String())
		default:
			fmt.Println(err)
		}
		os.Exit(64)
	}
}
