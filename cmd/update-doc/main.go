package main

import (
	"github.com/c9s/bbgo/pkg/cmd"
	"github.com/spf13/cobra/doc"
	"path"
	"runtime"
	"fmt"
	"log"
)

func main() {
	_, b, _, _ := runtime.Caller(0)
	root := path.Join(path.Dir(path.Dir(path.Dir(b))), "doc", "commands")
	fmt.Println(root)
	if err := doc.GenMarkdownTree(cmd.RootCmd, root); err != nil {
		log.Fatal(err)
	}
}
