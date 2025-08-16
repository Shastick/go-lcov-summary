package lcov

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummarize(t *testing.T) {
	// Test with sample.lcov
	file, err := os.Open("testdata/sample.lcov")
	require.NoError(t, err)
	defer file.Close()

	summary, err := Summarize(file)
	require.NoError(t, err)
	require.NotNil(t, summary)

	// Verify summary data
	assert.Equal(t, 2, summary.TotalFiles)
	assert.Equal(t, 9, summary.TotalLines)                   // 5 + 4
	assert.Equal(t, 6, summary.CoveredLines)                 // 3 + 3
	assert.InDelta(t, 66.67, summary.LineCoverageRate, 0.01) // 6/9 * 100

	// Verify summary statistics only (no individual file details)
}

func TestSummarizeComplex(t *testing.T) {
	// Test with complex.lcov
	file, err := os.Open("testdata/complex.lcov")
	require.NoError(t, err)
	defer file.Close()

	summary, err := Summarize(file)
	require.NoError(t, err)
	require.NotNil(t, summary)

	// Verify summary data
	assert.Equal(t, 3, summary.TotalFiles)
	assert.Equal(t, 15, summary.TotalLines)                  // 7 + 5 + 3
	assert.Equal(t, 11, summary.CoveredLines)                // 5 + 3 + 3
	assert.InDelta(t, 73.33, summary.LineCoverageRate, 0.01) // 11/15 * 100

	// Verify summary statistics only (no individual file details)
}

func TestParserParseRecord(t *testing.T) {
	parser := &Parser{}
	tests := []struct {
		name     string
		input    string
		expected *Record
		err      string
	}{
		// Existing valid cases
		{
			name:     "valid test name",
			input:    "TN:TestName",
			expected: &Record{Type: recordTestName, Value: "TestName"},
			err:      "",
		},
		{
			name:     "valid source file",
			input:    "SF:/path/to/file.go",
			expected: &Record{Type: recordSourceFile, Value: "/path/to/file.go"},
			err:      "",
		},
		{
			name:     "valid line data",
			input:    "DA:1,5",
			expected: &Record{Type: recordLineData, Value: "1,5"},
			err:      "",
		},
		{
			name:     "valid end of record",
			input:    "end_of_record",
			expected: &Record{Type: recordEndOfRecord, Value: ""},
			err:      "",
		},
		// New invalid cases
		{
			name:  "empty record",
			input: ":",
			err:   "invalid record format: :",
		},
		{
			name:  "missing value",
			input: "SF:",
			err:   "invalid record format: SF:",
		},
		{
			name:     "colon in value",
			input:    "DA:1:5",
			err:      "", // Should parse as DA with value "1:5"
			expected: &Record{Type: recordLineData, Value: "1:5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := parser.parseRecord(tt.input)
			if tt.err != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.err)
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, record)
			}
		})
	}
}

func TestParserParseLineData(t *testing.T) {
	parser := &Parser{}
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Existing valid cases
		{name: "valid line data", input: "1,5", expected: true},
		{name: "valid zero count", input: "10,0", expected: true},
		// New invalid cases
		{name: "missing comma", input: "1", expected: false},
		{name: "non-numeric line", input: "invalid,5", expected: false},
		{name: "non-numeric count", input: "1,invalid", expected: false},
		{name: "empty", input: "", expected: false},
		{name: "too many parts", input: "1,2,3", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, parser.isValidLineData(tt.input))
		})
	}
}

func TestSummarizeWithFunctionsAndBranches(t *testing.T) {
	// Test with function and branch coverage data
	file, err := os.Open("testdata/with_functions_and_branches.lcov")
	require.NoError(t, err)
	defer file.Close()

	summary, err := Summarize(file)
	require.NoError(t, err)
	require.NotNil(t, summary)

	// Verify summary data
	assert.Equal(t, 2, summary.TotalFiles)
	assert.Equal(t, 10, summary.TotalLines)                     // 6 + 4
	assert.Equal(t, 7, summary.CoveredLines)                    // 4 + 3
	assert.InDelta(t, 70.0, summary.LineCoverageRate, 0.01)     // 7/10 * 100
	assert.Equal(t, 4, summary.TotalFunctions)                  // 2 + 2
	assert.Equal(t, 3, summary.CoveredFunctions)                // 1 + 2 (functions with exec count > 0)
	assert.InDelta(t, 75.0, summary.FunctionCoverageRate, 0.01) // 3/4 * 100
	assert.Equal(t, 2, summary.TotalBranches)                   // 0 + 2
	assert.Equal(t, 2, summary.CoveredBranches)                 // 0 + 2
	assert.InDelta(t, 100.0, summary.BranchCoverageRate, 0.01)  // 2/2 * 100

	// Verify summary statistics only (no individual file details)
}

func TestParserParseFunctionName(t *testing.T) {
	parser := &Parser{}
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Existing valid cases
		{name: "valid function name", input: "1,main", expected: true},
		{name: "valid helper function", input: "10,helper", expected: true},
		// New invalid cases
		{name: "missing comma", input: "1main", expected: false},
		{name: "non-numeric line", input: "invalid,main", expected: false},
		{name: "empty name", input: "1,", expected: false},
		{name: "empty input", input: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, parser.isValidFunctionName(tt.input))
		})
	}
}

func TestParserParseBranchData(t *testing.T) {
	parser := &Parser{}
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Existing valid cases
		{name: "valid branch data", input: "1,0,0,1", expected: true},
		{name: "valid zero count", input: "5,1,2,0", expected: true},
		{name: "valid dash count", input: "10,2,1,-", expected: true},
		// New invalid cases
		{name: "missing parts", input: "1,0,0", expected: false},
		{name: "non-numeric line", input: "invalid,0,0,1", expected: false},
		{name: "non-numeric block", input: "1,invalid,0,1", expected: false},
		{name: "non-numeric branch", input: "1,0,invalid,1", expected: false},
		{name: "invalid count", input: "1,0,0,invalid", expected: false},
		{name: "empty input", input: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, parser.isValidBranchData(tt.input))
		})
	}
}

func TestSummarizeErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "line data without source file",
			input: "DA:1,5\nend_of_record",
			err:   "line data without source file",
		},
		{
			name:  "lines found without source file",
			input: "LF:10\nend_of_record",
			err:   "lines found without source file",
		},
		{
			name:  "lines hit without source file",
			input: "LH:5\nend_of_record",
			err:   "lines hit without source file",
		},
		{
			name:  "function name without source file",
			input: "FN:1,main\nend_of_record",
			err:   "function name without source file",
		},
		{
			name:  "function data without source file",
			input: "FNDA:1,main\nend_of_record",
			err:   "function data without source file",
		},
		{
			name:  "branch data without source file",
			input: "BRDA:1,0,0,1\nend_of_record",
			err:   "branch data without source file",
		},
		{
			name:  "branch found without source file",
			input: "BRF:2\nend_of_record",
			err:   "branch found without source file",
		},
		{
			name:  "branch hit without source file",
			input: "BRH:1\nend_of_record",
			err:   "branch hit without source file",
		},
		{
			name:  "invalid lines found value",
			input: "SF:/path/to/file.go\nLF:invalid\nend_of_record",
			err:   "invalid lines found value: invalid",
		},
		{
			name:  "invalid lines hit value",
			input: "SF:/path/to/file.go\nLH:invalid\nend_of_record",
			err:   "invalid lines hit value: invalid",
		},
		{
			name:  "invalid branch found value",
			input: "SF:/path/to/file.go\nBRF:invalid\nend_of_record",
			err:   "invalid branches found value: invalid",
		},
		{
			name:  "invalid branch hit value",
			input: "SF:/path/to/file.go\nBRH:invalid\nend_of_record",
			err:   "invalid branches hit value: invalid",
		},
		{
			name:  "invalid line data format",
			input: "SF:/path/to/file.go\nDA:1\nend_of_record",
			err:   "invalid line data format: 1",
		},
		{
			name:  "invalid function name format",
			input: "SF:/path/to/file.go\nFN:invalid\nend_of_record",
			err:   "invalid function name format: invalid",
		},
		{
			name:  "invalid branch data format",
			input: "SF:/path/to/file.go\nBRDA:1,0,0\nend_of_record",
			err:   "invalid branch data format: 1,0,0",
		},
		{
			name:  "invalid record type",
			input: "FN:\nend_of_record",
			err:   "failed to parse line 'FN:': invalid record format: FN:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			summary, err := Summarize(reader)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
			assert.Nil(t, summary)
		})
	}
}

type failingReader struct{}

func (r *failingReader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("simulated read error")
}

func TestSummarizeScannerError(t *testing.T) {
	parser := NewParser(&failingReader{})
	summary, err := parser.Parse()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated read error")
	assert.Nil(t, summary)
}
