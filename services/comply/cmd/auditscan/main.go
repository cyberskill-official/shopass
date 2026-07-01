package main

import (
	"flag"
	"fmt"
	"os"

	"shopass/services/comply/internal/audit"
)

func main() {
	root := flag.String("root", ".", "Root directory to scan")
	count := flag.Bool("count", false, "Print only the number of findings")
	flag.Parse()

	findings, err := audit.Scan(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
		os.Exit(1)
	}

	if *count {
		fmt.Println(len(findings))
		return
	}

	for _, f := range findings {
		fmt.Printf("%s:%d [%s] %s\n", f.File, f.Line, f.Rule, f.Hint)
	}

	if len(findings) > 0 {
		os.Exit(1)
	}
}
