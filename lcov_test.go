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

	// Verify file records
	assert.Len(t, summary.Files, 2)

	// First file
	assert.Equal(t, "", summary.Files[0].TestName)
	assert.Equal(t, "/path/to/source/file1.go", summary.Files[0].SourceFile)
	assert.Equal(t, 5, summary.Files[0].LinesFound)
	assert.Equal(t, 3, summary.Files[0].LinesHit)
	assert.Len(t, summary.Files[0].Lines, 5)

	// Second file
	assert.Equal(t, "", summary.Files[1].TestName)
	assert.Equal(t, "/path/to/source/file2.go", summary.Files[1].SourceFile)
	assert.Equal(t, 4, summary.Files[1].LinesFound)
	assert.Equal(t, 3, summary.Files[1].LinesHit)
	assert.Len(t, summary.Files[1].Lines, 4)
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

	// Verify file records
	assert.Len(t, summary.Files, 3)

	// First file (TestSuite1)
	assert.Equal(t, "TestSuite1", summary.Files[0].TestName)
	assert.Equal(t, "/path/to/source/main.go", summary.Files[0].SourceFile)
	assert.Equal(t, 7, summary.Files[0].LinesFound)
	assert.Equal(t, 5, summary.Files[0].LinesHit)
	assert.Len(t, summary.Files[0].Lines, 7)

	// Second file (TestSuite2)
	assert.Equal(t, "TestSuite2", summary.Files[1].TestName)
	assert.Equal(t, "/path/to/source/utils.go", summary.Files[1].SourceFile)
	assert.Equal(t, 5, summary.Files[1].LinesFound)
	assert.Equal(t, 3, summary.Files[1].LinesHit)
	assert.Len(t, summary.Files[1].Lines, 5)

	// Third file (TestSuite3)
	assert.Equal(t, "TestSuite3", summary.Files[2].TestName)
	assert.Equal(t, "/path/to/source/helper.go", summary.Files[2].SourceFile)
	assert.Equal(t, 3, summary.Files[2].LinesFound)
	assert.Equal(t, 3, summary.Files[2].LinesHit)
	assert.Len(t, summary.Files[2].Lines, 3)
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
	lineData, err := parser.parseLineData("1,5")
	assert.NoError(t, err)
	assert.Equal(t, 1, lineData.LineNumber)
	assert.Equal(t, 5, lineData.ExecutionCount)

	lineData, err = parser.parseLineData("10,0")
	assert.NoError(t, err)
	assert.Equal(t, 10, lineData.LineNumber)
	assert.Equal(t, 0, lineData.ExecutionCount)

	// Test invalid line data
	_, err = parser.parseLineData("invalid")
	assert.Error(t, err)

	_, err = parser.parseLineData("1,invalid")
	assert.Error(t, err)

	_, err = parser.parseLineData("invalid,5")
	assert.Error(t, err)
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

	// Verify file records
	assert.Len(t, summary.Files, 2)

	// First file (main.go)
	assert.Equal(t, "TestSuite", summary.Files[0].TestName)
	assert.Equal(t, "/path/to/source/main.go", summary.Files[0].SourceFile)
	assert.Equal(t, 6, summary.Files[0].LinesFound)
	assert.Equal(t, 4, summary.Files[0].LinesHit)
	assert.Len(t, summary.Files[0].Lines, 6)
	assert.Len(t, summary.Files[0].Functions, 2)
	assert.Equal(t, 1, summary.Files[0].FunctionsHit)
	assert.Len(t, summary.Files[0].Branches, 0)
	assert.Equal(t, 0, summary.Files[0].BranchesFound)
	assert.Equal(t, 0, summary.Files[0].BranchesHit)

	// Second file (utils.go)
	assert.Equal(t, "TestSuite", summary.Files[1].TestName)
	assert.Equal(t, "/path/to/source/utils.go", summary.Files[1].SourceFile)
	assert.Equal(t, 4, summary.Files[1].LinesFound)
	assert.Equal(t, 3, summary.Files[1].LinesHit)
	assert.Len(t, summary.Files[1].Lines, 4)
	assert.Len(t, summary.Files[1].Functions, 2)
	assert.Equal(t, 2, summary.Files[1].FunctionsHit)
	assert.Len(t, summary.Files[1].Branches, 4)
	assert.Equal(t, 2, summary.Files[1].BranchesFound)
	assert.Equal(t, 2, summary.Files[1].BranchesHit)
}

func TestParserParseFunctionName(t *testing.T) {
	parser := &Parser{}

	// Test valid function name
	functionData, err := parser.parseFunctionName("1,main")
	assert.NoError(t, err)
	assert.Equal(t, 1, functionData.LineNumber)
	assert.Equal(t, "main", functionData.Name)

	functionData, err = parser.parseFunctionName("10,helper")
	assert.NoError(t, err)
	assert.Equal(t, 10, functionData.LineNumber)
	assert.Equal(t, "helper", functionData.Name)

	// Test invalid function name
	_, err = parser.parseFunctionName("invalid")
	assert.Error(t, err)

	_, err = parser.parseFunctionName("1")
	assert.Error(t, err)
}

func TestParserParseBranchData(t *testing.T) {
	parser := &Parser{}

	// Test valid branch data
	branchData, err := parser.parseBranchData("1,0,0,1")
	assert.NoError(t, err)
	assert.Equal(t, 1, branchData.LineNumber)
	assert.Equal(t, 0, branchData.BlockNumber)
	assert.Equal(t, 0, branchData.BranchNumber)
	assert.Equal(t, 1, branchData.ExecutionCount)

	branchData, err = parser.parseBranchData("5,1,2,0")
	assert.NoError(t, err)
	assert.Equal(t, 5, branchData.LineNumber)
	assert.Equal(t, 1, branchData.BlockNumber)
	assert.Equal(t, 2, branchData.BranchNumber)
	assert.Equal(t, 0, branchData.ExecutionCount)

	branchData, err = parser.parseBranchData("10,2,1,-")
	assert.NoError(t, err)
	assert.Equal(t, 10, branchData.LineNumber)
	assert.Equal(t, 2, branchData.BlockNumber)
	assert.Equal(t, 1, branchData.BranchNumber)
	assert.Equal(t, 0, branchData.ExecutionCount) // "-" should be parsed as 0

	// Test invalid branch data
	_, err = parser.parseBranchData("invalid")
	assert.Error(t, err)

	_, err = parser.parseBranchData("1,0,0")
	assert.Error(t, err)

	_, err = parser.parseBranchData("1,0,0,invalid")
	assert.Error(t, err)
}
