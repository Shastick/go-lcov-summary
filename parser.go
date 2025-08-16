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
	summary := &Summary{}

	// Current file counters
	var currentFileLinesFound, currentFileLinesHit int
	var currentFileFunctions, currentFileFunctionsHit int
	var currentFileBranchesFound, currentFileBranchesHit int
	var inFile bool

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
			// Test name - just continue, we don't need to store it

		case RecordSourceFile:
			// Start of a new file
			inFile = true
			currentFileLinesFound = 0
			currentFileLinesHit = 0
			currentFileFunctions = 0
			currentFileFunctionsHit = 0
			currentFileBranchesFound = 0
			currentFileBranchesHit = 0

		case RecordLineData:
			if !inFile {
				return nil, fmt.Errorf("line data without source file")
			}
			// We don't need to store individual line data, just validate the format
			if !p.isValidLineData(record.Value) {
				return nil, fmt.Errorf("invalid line data format: %s", record.Value)
			}

		case RecordLinesFound:
			if !inFile {
				return nil, fmt.Errorf("lines found without source file")
			}
			linesFound, err := strconv.Atoi(record.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid lines found value: %s", record.Value)
			}
			currentFileLinesFound = linesFound

		case RecordLinesHit:
			if !inFile {
				return nil, fmt.Errorf("lines hit without source file")
			}
			linesHit, err := strconv.Atoi(record.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid lines hit value: %s", record.Value)
			}
			currentFileLinesHit = linesHit

		case RecordFunctionName:
			if !inFile {
				return nil, fmt.Errorf("function name without source file")
			}
			// We don't need to store function data, just validate and count
			if !p.isValidFunctionName(record.Value) {
				return nil, fmt.Errorf("invalid function name format: %s", record.Value)
			}
			currentFileFunctions++

		case RecordFunctionData:
			if !inFile {
				return nil, fmt.Errorf("function data without source file")
			}
			// FNDA records are matched with FN records by name
			// For simplicity, we'll just count functions that were executed
			parts := strings.Split(record.Value, ",")
			if len(parts) == 2 {
				execCount, err := strconv.Atoi(parts[0])
				if err == nil && execCount > 0 {
					currentFileFunctionsHit++
				}
			}

		case RecordBranchData:
			if !inFile {
				return nil, fmt.Errorf("branch data without source file")
			}
			// We don't need to store branch data, just validate the format
			if !p.isValidBranchData(record.Value) {
				return nil, fmt.Errorf("invalid branch data format: %s", record.Value)
			}

		case RecordBranchFound:
			if !inFile {
				return nil, fmt.Errorf("branch found without source file")
			}
			branchesFound, err := strconv.Atoi(record.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid branches found value: %s", record.Value)
			}
			currentFileBranchesFound = branchesFound

		case RecordBranchHit:
			if !inFile {
				return nil, fmt.Errorf("branch hit without source file")
			}
			branchesHit, err := strconv.Atoi(record.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid branches hit value: %s", record.Value)
			}
			currentFileBranchesHit = branchesHit

		case RecordEndOfRecord:
			if inFile {
				// Add current file's data to totals
				summary.TotalFiles++
				summary.TotalLines += currentFileLinesFound
				summary.CoveredLines += currentFileLinesHit
				summary.TotalFunctions += currentFileFunctions
				summary.CoveredFunctions += currentFileFunctionsHit
				summary.TotalBranches += currentFileBranchesFound
				summary.CoveredBranches += currentFileBranchesHit
				inFile = false
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

// isValidLineData validates a line data record (DA:line,count)
func (p *Parser) isValidLineData(value string) bool {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return false
	}

	_, err1 := strconv.Atoi(parts[0])
	_, err2 := strconv.Atoi(parts[1])
	return err1 == nil && err2 == nil
}

// isValidFunctionName validates a function name record (FN:line,name)
func (p *Parser) isValidFunctionName(value string) bool {
	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 {
		return false
	}

	_, err := strconv.Atoi(parts[0])
	return err == nil
}

// isValidBranchData validates a branch data record (BRDA:line,block,branch,count)
func (p *Parser) isValidBranchData(value string) bool {
	parts := strings.Split(value, ",")
	if len(parts) != 4 {
		return false
	}

	_, err1 := strconv.Atoi(parts[0])
	_, err2 := strconv.Atoi(parts[1])
	_, err3 := strconv.Atoi(parts[2])

	// The fourth part can be a number or "-"
	var err4 error
	if parts[3] != "-" {
		_, err4 = strconv.Atoi(parts[3])
	}

	return err1 == nil && err2 == nil && err3 == nil && (parts[3] == "-" || err4 == nil)
}
