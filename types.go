package lcov

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// RecordType represents the type of LCOV record
type RecordType string

const (
	RecordTestName     RecordType = "TN"
	RecordSourceFile   RecordType = "SF"
	RecordLineData     RecordType = "DA"
	RecordLinesFound   RecordType = "LF"
	RecordLinesHit     RecordType = "LH"
	RecordEndOfRecord  RecordType = "end_of_record"
	RecordFunctionName RecordType = "FN"
	RecordFunctionData RecordType = "FNDA"
	RecordBranchData   RecordType = "BRDA"
	RecordBranchFound  RecordType = "BRF"
	RecordBranchHit    RecordType = "BRH"
)

// LineData represents a single line's coverage data
type LineData struct {
	LineNumber     int
	ExecutionCount int
}

// FunctionData represents function coverage data
type FunctionData struct {
	LineNumber int
	Name       string
}

// BranchData represents branch coverage data
type BranchData struct {
	LineNumber     int
	BlockNumber    int
	BranchNumber   int
	ExecutionCount int
}

// FileRecord represents the coverage data for a single source file
type FileRecord struct {
	TestName      string
	SourceFile    string
	Lines         []LineData
	LinesFound    int
	LinesHit      int
	Functions     []FunctionData
	FunctionsHit  int
	Branches      []BranchData
	BranchesFound int
	BranchesHit   int
}

// Summary represents the overall coverage summary
type Summary struct {
	TotalFiles           int
	TotalLines           int
	CoveredLines         int
	LineCoverageRate     float64
	TotalFunctions       int
	CoveredFunctions     int
	FunctionCoverageRate float64
	TotalBranches        int
	CoveredBranches      int
	BranchCoverageRate   float64
	Files                []FileRecord
}

// Parser represents an LCOV file parser
type Parser struct {
	scanner *bufio.Scanner
}

// NewParser creates a new LCOV parser
func NewParser(reader io.Reader) *Parser {
	return &Parser{
		scanner: bufio.NewScanner(reader),
	}
}

// Parse reads and parses the entire LCOV file
func (p *Parser) Parse() (*Summary, error) {
	summary := &Summary{
		Files: make([]FileRecord, 0),
	}

	var currentFile *FileRecord

	for p.scanner.Scan() {
		line := strings.TrimSpace(p.scanner.Text())
		if line == "" {
			continue
		}

		record, err := p.parseRecord(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line '%s': %w", line, err)
		}

		switch record.Type {
		case RecordTestName:
			// Start a new file record if we don't have one
			if currentFile == nil {
				currentFile = &FileRecord{
					TestName:  record.Value,
					Lines:     make([]LineData, 0),
					Functions: make([]FunctionData, 0),
					Branches:  make([]BranchData, 0),
				}
			} else {
				currentFile.TestName = record.Value
			}

		case RecordSourceFile:
			if currentFile == nil {
				currentFile = &FileRecord{
					Lines:     make([]LineData, 0),
					Functions: make([]FunctionData, 0),
					Branches:  make([]BranchData, 0),
				}
			}
			currentFile.SourceFile = record.Value

		case RecordLineData:
			if currentFile == nil {
				return nil, fmt.Errorf("line data without source file")
			}
			lineData, err := p.parseLineData(record.Value)
			if err != nil {
				return nil, err
			}
			currentFile.Lines = append(currentFile.Lines, lineData)

		case RecordLinesFound:
			if currentFile == nil {
				return nil, fmt.Errorf("lines found without source file")
			}
			linesFound, err := strconv.Atoi(record.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid lines found value: %s", record.Value)
			}
			currentFile.LinesFound = linesFound

		case RecordLinesHit:
			if currentFile == nil {
				return nil, fmt.Errorf("lines hit without source file")
			}
			linesHit, err := strconv.Atoi(record.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid lines hit value: %s", record.Value)
			}
			currentFile.LinesHit = linesHit

		case RecordFunctionName:
			if currentFile == nil {
				return nil, fmt.Errorf("function name without source file")
			}
			functionData, err := p.parseFunctionName(record.Value)
			if err != nil {
				return nil, err
			}
			currentFile.Functions = append(currentFile.Functions, functionData)

		case RecordFunctionData:
			if currentFile == nil {
				return nil, fmt.Errorf("function data without source file")
			}
			// FNDA records are matched with FN records by name
			// For simplicity, we'll just count functions that were executed
			parts := strings.Split(record.Value, ",")
			if len(parts) == 2 {
				execCount, err := strconv.Atoi(parts[0])
				if err == nil && execCount > 0 {
					currentFile.FunctionsHit++
				}
			}

		case RecordBranchData:
			if currentFile == nil {
				return nil, fmt.Errorf("branch data without source file")
			}
			branchData, err := p.parseBranchData(record.Value)
			if err != nil {
				return nil, err
			}
			currentFile.Branches = append(currentFile.Branches, branchData)

		case RecordBranchFound:
			if currentFile == nil {
				return nil, fmt.Errorf("branch found without source file")
			}
			branchesFound, err := strconv.Atoi(record.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid branches found value: %s", record.Value)
			}
			currentFile.BranchesFound = branchesFound

		case RecordBranchHit:
			if currentFile == nil {
				return nil, fmt.Errorf("branch hit without source file")
			}
			branchesHit, err := strconv.Atoi(record.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid branches hit value: %s", record.Value)
			}
			currentFile.BranchesHit = branchesHit

		case RecordEndOfRecord:
			if currentFile != nil {
				summary.Files = append(summary.Files, *currentFile)
				summary.TotalFiles++
				summary.TotalLines += currentFile.LinesFound
				summary.CoveredLines += currentFile.LinesHit
				summary.TotalFunctions += len(currentFile.Functions)
				summary.CoveredFunctions += currentFile.FunctionsHit
				summary.TotalBranches += currentFile.BranchesFound
				summary.CoveredBranches += currentFile.BranchesHit
				currentFile = nil
			}
		}
	}

	// Calculate coverage rates
	if summary.TotalLines > 0 {
		summary.LineCoverageRate = float64(summary.CoveredLines) / float64(summary.TotalLines) * 100
	}
	if summary.TotalFunctions > 0 {
		summary.FunctionCoverageRate = float64(summary.CoveredFunctions) / float64(summary.TotalFunctions) * 100
	}
	if summary.TotalBranches > 0 {
		summary.BranchCoverageRate = float64(summary.CoveredBranches) / float64(summary.TotalBranches) * 100
	}

	return summary, p.scanner.Err()
}

// Record represents a parsed LCOV record
type Record struct {
	Type  RecordType
	Value string
}

// parseRecord parses a single line into a Record
func (p *Parser) parseRecord(line string) (*Record, error) {
	if line == "end_of_record" {
		return &Record{Type: RecordEndOfRecord, Value: ""}, nil
	}

	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid record format: %s", line)
	}

	recordType := RecordType(parts[0])
	value := parts[1]

	return &Record{Type: recordType, Value: value}, nil
}

// parseLineData parses a line data record (DA:line,count)
func (p *Parser) parseLineData(value string) (LineData, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return LineData{}, fmt.Errorf("invalid line data format: %s", value)
	}

	lineNumber, err := strconv.Atoi(parts[0])
	if err != nil {
		return LineData{}, fmt.Errorf("invalid line number: %s", parts[0])
	}

	executionCount, err := strconv.Atoi(parts[1])
	if err != nil {
		return LineData{}, fmt.Errorf("invalid execution count: %s", parts[1])
	}

	return LineData{
		LineNumber:     lineNumber,
		ExecutionCount: executionCount,
	}, nil
}

// parseFunctionName parses a function name record (FN:line,name)
func (p *Parser) parseFunctionName(value string) (FunctionData, error) {
	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 {
		return FunctionData{}, fmt.Errorf("invalid function name format: %s", value)
	}

	lineNumber, err := strconv.Atoi(parts[0])
	if err != nil {
		return FunctionData{}, fmt.Errorf("invalid function line number: %s", parts[0])
	}

	return FunctionData{
		LineNumber: lineNumber,
		Name:       parts[1],
	}, nil
}

// parseBranchData parses a branch data record (BRDA:line,block,branch,count)
func (p *Parser) parseBranchData(value string) (BranchData, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 4 {
		return BranchData{}, fmt.Errorf("invalid branch data format: %s", value)
	}

	lineNumber, err := strconv.Atoi(parts[0])
	if err != nil {
		return BranchData{}, fmt.Errorf("invalid branch line number: %s", parts[0])
	}

	blockNumber, err := strconv.Atoi(parts[1])
	if err != nil {
		return BranchData{}, fmt.Errorf("invalid branch block number: %s", parts[1])
	}

	branchNumber, err := strconv.Atoi(parts[2])
	if err != nil {
		return BranchData{}, fmt.Errorf("invalid branch number: %s", parts[2])
	}

	executionCount := 0
	if parts[3] != "-" {
		executionCount, err = strconv.Atoi(parts[3])
		if err != nil {
			return BranchData{}, fmt.Errorf("invalid branch execution count: %s", parts[3])
		}
	}

	return BranchData{
		LineNumber:     lineNumber,
		BlockNumber:    blockNumber,
		BranchNumber:   branchNumber,
		ExecutionCount: executionCount,
	}, nil
}
