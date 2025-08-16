# go-lcov-summary

Native Go CLI and tiny library aimed at computing summaries of LCOV files, as a replacement to using the original [lcov](https://github.com/linux-test-project/lcov) CLI directly, with the `lcov --sumary` option.

## Usage

### CLI

```bash
# Mind the lowercased 's' in the repo name!
go install github.com/shastick/go-lcov-summary/cmd/go-lcov-summary@latest
go-lcov-summary <lcov-file>
```

Which should output something like:
```
Summary coverage rate:
  source files: 2
  lines.......: 70.0% (7 of 10 lines)
  functions...: 75.0% (3 of 4 functions)
  branches....: 100.0% (2 of 2 branches)
```

### Library

```
# Mind the lowercased 's' in the repo name!
go get github.com/shastick/go-lcov-summary@latest

# In your code:

file, err := os.Open("coverage.lcov")
summary, err := lcov.Summarize(file)
```

with `summary` being a `lcov.Summary` struct with the following fields:

```
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
```

## Performance

go-lcov-summary is built with performance in mind, but no particular performance tests or benchmarks have been run.