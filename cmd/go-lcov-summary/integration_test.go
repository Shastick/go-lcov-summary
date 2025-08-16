package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationLCOVSummary(t *testing.T) {
	// Define test cases with LCOV files
	testCases := []struct {
		name     string
		input    string
		hasError bool
	}{
		{name: "sample", input: "testdata/sample.lcov"},
		{name: "complex", input: "testdata/complex.lcov"},
		{name: "functions_and_branches", input: "testdata/with_functions_and_branches.lcov"},
		{name: "invalid", input: "testdata/invalid.lcov", hasError: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run go-lcov-summary
			goOutput, goErr := runGoLCOVSummary(t, tc.input)
			if tc.hasError {
				require.Error(t, goErr, "go-lcov-summary should fail for invalid input")
				return
			}
			require.NoError(t, goErr, "go-lcov-summary should not fail")

			// Run lcov --summary
			lcovOutput, lcovErr := runLCOVSummary(t, tc.input)
			require.NoError(t, lcovErr, "lcov --summary should not fail")

			// Parse and compare outputs
			goMetrics := parseGoLCOVSummaryOutput(goOutput)
			lcovMetrics := parseLCOVSummaryOutput(lcovOutput)

			if !tc.hasError {
				require.Greater(t, goMetrics.TotalLines, 0, "Total lines should be greater than 0")
			}

			// Compare metrics
			assert.InDelta(t, goMetrics.LineCoverageRate, lcovMetrics.LineCoverageRate, 0.01, "Line coverage rate mismatch")
			assert.Equal(t, goMetrics.TotalLines, lcovMetrics.TotalLines, "Total lines mismatch")
			assert.Equal(t, goMetrics.CoveredLines, lcovMetrics.CoveredLines, "Covered lines mismatch")
			assert.InDelta(t, goMetrics.FunctionCoverageRate, lcovMetrics.FunctionCoverageRate, 0.01, "Function coverage rate mismatch")
			assert.Equal(t, goMetrics.TotalFunctions, lcovMetrics.TotalFunctions, "Total functions mismatch")
			assert.Equal(t, goMetrics.CoveredFunctions, lcovMetrics.CoveredFunctions, "Covered functions mismatch")
		})
	}
}

// Metrics holds parsed summary data
type Metrics struct {
	LineCoverageRate     float64
	TotalLines           int
	CoveredLines         int
	FunctionCoverageRate float64
	TotalFunctions       int
	CoveredFunctions     int
}

// runGoLCOVSummary runs go-lcov-summary in a Docker container
func runGoLCOVSummary(t *testing.T, input string) (string, error) {
	args := []string{"run", "--rm", "-v", fmt.Sprintf("%s:/app/testdata", filepath.Dir(input)), "go-lcov-summary:latest", "/app/testdata/" + filepath.Base(input)}

	cmd := exec.Command("docker", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	t.Log("GO OUT"+": ", out.String())
	return out.String(), err
}

// runLCOVSummary runs lcov --summary in a Docker container
func runLCOVSummary(t *testing.T, input string) (string, error) {
	// lcov --summary can't read from stdin
	args := []string{"run", "--rm", "-v", fmt.Sprintf("%s:/app/testdata", filepath.Dir(input)), "lcov-test:latest", "lcov", "--summary", "/app/testdata/" + filepath.Base(input)}

	cmd := exec.Command("docker", args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	t.Log("LCOV OUT"+": ", out.String())
	return out.String(), err
}

// parseGoLCOVSummaryOutput parses go-lcov-summary output
func parseGoLCOVSummaryOutput(output string) Metrics {
	var m Metrics
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Note we ignore source files and branch counts as they are only provided by lcov on OS X, not Ubuntu & co
		if strings.HasPrefix(line, "lines.......:") {
			var perc string
			fmt.Sscanf(line, "lines.......: %s (%d of %d lines)", &perc, &m.CoveredLines, &m.TotalLines)
			perc = strings.TrimSuffix(perc, "%")
			m.LineCoverageRate, _ = strconv.ParseFloat(perc, 64)
		} else if strings.HasPrefix(line, "functions...:") && !strings.Contains(line, "no data found") {
			var perc string
			fmt.Sscanf(line, "functions...: %s (%d of %d functions)", &perc, &m.CoveredFunctions, &m.TotalFunctions)
			perc = strings.TrimSuffix(perc, "%")
			m.FunctionCoverageRate, _ = strconv.ParseFloat(perc, 64)
		}
	}
	return m
}

// parseLCOVSummaryOutput parses lcov --summary output
func parseLCOVSummaryOutput(output string) Metrics {
	var m Metrics
	lines := strings.Split(output, "\n")
	re := regexp.MustCompile(`(\d+\.\d+%)\s+\((\d+) of (\d+) (lines|functions|branches)\)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if matches := re.FindStringSubmatch(line); matches != nil {
			perc, _ := strconv.ParseFloat(strings.TrimSuffix(matches[1], "%"), 64)
			covered, _ := strconv.Atoi(matches[2])
			total, _ := strconv.Atoi(matches[3])
			switch matches[4] {
			case "lines":
				m.LineCoverageRate = perc
				m.CoveredLines = covered
				m.TotalLines = total
			case "functions":
				m.FunctionCoverageRate = perc
				m.CoveredFunctions = covered
				m.TotalFunctions = total
			}
		}
	}

	return m
}
