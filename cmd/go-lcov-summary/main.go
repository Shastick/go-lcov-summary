package main

import (
	"fmt"
	"github.com/shastick/go-lcov-summary"
	"io"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <lcov-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s - (read from stdin)\n", os.Args[0])
		os.Exit(1)
	}

	var reader io.Reader

	if os.Args[1] == "-" {
		// Read from stdin
		reader = os.Stdin
	} else {
		// Read from file
		file, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		reader = file
	}

	summary, err := lcov.Summarize(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing LCOV file: %v\n", err)
		os.Exit(1)
	}

	// Display summary
	displaySummary(summary)
}

func displaySummary(summary *lcov.Summary) {
	fmt.Println("Summary coverage rate:")
	fmt.Printf("  source files: %d\n", summary.TotalFiles)
	fmt.Printf("  lines.......: %.1f%% (%d of %d lines)\n",
		summary.LineCoverageRate, summary.CoveredLines, summary.TotalLines)

	if summary.TotalFunctions > 0 {
		fmt.Printf("  functions...: %.1f%% (%d of %d functions)\n",
			summary.FunctionCoverageRate, summary.CoveredFunctions, summary.TotalFunctions)
	} else {
		fmt.Println("  functions...: no data found")
	}

	if summary.TotalBranches > 0 {
		fmt.Printf("  branches....: %.1f%% (%d of %d branches)\n",
			summary.BranchCoverageRate, summary.CoveredBranches, summary.TotalBranches)
	} else {
		fmt.Println("  branches....: no data found")
	}
}
