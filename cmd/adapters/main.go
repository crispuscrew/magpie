// Command adapters regenerates host rule files from AGENTS.md; --check fails on drift.
package main

import (
	"fmt"
	"os"

	"github.com/crispuscrew/magpie/internal/adapters"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(os.Args) > 1 && os.Args[1] == "--check" {
		drifted, err := adapters.Check(root)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if len(drifted) > 0 {
			fmt.Fprintln(os.Stderr, "drifted from AGENTS.md (run: make adapters):")
			for _, path := range drifted {
				fmt.Fprintln(os.Stderr, "  "+path)
			}
			os.Exit(1)
		}
		fmt.Printf("%d adapters match AGENTS.md\n", len(adapters.Files))
		return
	}
	if err := adapters.Build(root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %d adapters\n", len(adapters.Files))
}
