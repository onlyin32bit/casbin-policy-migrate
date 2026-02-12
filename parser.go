package casbinMigrate

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// OperationType represents the type of operation (Add or Remove).
type OperationType int

const (
	OperationAdd OperationType = iota
	OperationRemove
)

// Operation represents a single policy change operation.
type Operation struct {
	Type  OperationType
	Sec   string   // "p" or "g"
	PType string   // "p", "g", "g2", etc.
	Rule  []string // The policy rule content
}

// ParseMigrationFile parses a CSV migration file and returns a list of operations.
func ParseMigrationFile(filePath string) ([]Operation, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ParseReader(f)
}

// ParseReader parses migration content from an io.Reader.
func ParseReader(r io.Reader) ([]Operation, error) {
	var ops []Operation
	scanner := bufio.NewScanner(r)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			continue
		}

		csvReader := csv.NewReader(strings.NewReader(line))
		record, err := csvReader.Read()
		if err != nil {
			return nil, fmt.Errorf("line %d: error parsing CSV: %w", lineNum, err)
		}

		if len(record) == 0 {
			continue
		}

		// Determine operation type and ptype
		firstCol := strings.TrimSpace(record[0])
		opType := OperationAdd

		if strings.HasPrefix(firstCol, "-") {
			opType = OperationRemove
			firstCol = strings.TrimPrefix(firstCol, "-")
		}

		ptype := firstCol
		// Determine section based on ptype (convention: p* -> p, g* -> g)
		sec := "p"
		if strings.HasPrefix(ptype, "g") {
			sec = "g"
		}

		// The rest of the record is the rule
		rule := make([]string, 0, len(record)-1)
		for _, col := range record[1:] {
			rule = append(rule, strings.TrimSpace(col))
		}

		ops = append(ops, Operation{
			Type:  opType,
			Sec:   sec,
			PType: ptype,
			Rule:  rule,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ops, nil
}
