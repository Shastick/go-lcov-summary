package lcov

import (
	"os"
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

	// Test valid records
	record, err := parser.parseRecord("TN:TestName")
	assert.NoError(t, err)
	assert.Equal(t, RecordTestName, record.Type)
	assert.Equal(t, "TestName", record.Value)

	record, err = parser.parseRecord("SF:/path/to/file.go")
	assert.NoError(t, err)
	assert.Equal(t, RecordSourceFile, record.Type)
	assert.Equal(t, "/path/to/file.go", record.Value)

	record, err = parser.parseRecord("DA:1,5")
	assert.NoError(t, err)
	assert.Equal(t, RecordLineData, record.Type)
	assert.Equal(t, "1,5", record.Value)

	record, err = parser.parseRecord("end_of_record")
	assert.NoError(t, err)
	assert.Equal(t, RecordEndOfRecord, record.Type)
	assert.Equal(t, "", record.Value)

	// Test invalid records
	_, err = parser.parseRecord("invalid")
	assert.Error(t, err)
}

func TestParserParseLineData(t *testing.T) {
	parser := &Parser{}

	// Test valid line data
	assert.True(t, parser.isValidLineData("1,5"))
	assert.True(t, parser.isValidLineData("10,0"))

	// Test invalid line data
	assert.False(t, parser.isValidLineData("invalid"))
	assert.False(t, parser.isValidLineData("1,invalid"))
	assert.False(t, parser.isValidLineData("invalid,5"))
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

	// Test valid function name
	assert.True(t, parser.isValidFunctionName("1,main"))
	assert.True(t, parser.isValidFunctionName("10,helper"))

	// Test invalid function name
	assert.False(t, parser.isValidFunctionName("invalid"))
	assert.False(t, parser.isValidFunctionName("1"))
}

func TestParserParseBranchData(t *testing.T) {
	parser := &Parser{}

	// Test valid branch data
	assert.True(t, parser.isValidBranchData("1,0,0,1"))
	assert.True(t, parser.isValidBranchData("5,1,2,0"))
	assert.True(t, parser.isValidBranchData("10,2,1,-"))

	// Test invalid branch data
	assert.False(t, parser.isValidBranchData("invalid"))
	assert.False(t, parser.isValidBranchData("1,0,0"))
	assert.False(t, parser.isValidBranchData("1,0,0,invalid"))
}
