package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/samber/lo"
)

type table struct {
	fieldWidths map[string]int
	rows        []map[string]any
}

func newTable() *table {
	return &table{
		fieldWidths: map[string]int{},
	}
}

func (t *table) addRow(row map[string]any) {
	for key, val := range row {
		val := fmt.Sprintf("%v", val)
		val = strings.ReplaceAll(val, "\n", "")
		val = strings.ReplaceAll(val, "\t", "        ")
		t.fieldWidths[key] = max(t.fieldWidths[key], len(val))
	}

	t.rows = append(t.rows, row)
}

func (t *table) print() {
	for fieldName, width := range t.fieldWidths {
		if width == 0 {
			delete(t.fieldWidths, fieldName)
			continue
		}

		t.fieldWidths[fieldName] = max(t.fieldWidths[fieldName], len(fmt.Sprintf("%v", fieldName)))
	}

	fields := lo.Keys(t.fieldWidths)

	sort.Slice(fields, func(i, j int) bool {
		order := map[string]int{
			"timestamp":  -5,
			"level":      -4,
			"filename":   -3,
			"linenumber": -2,
			"msg":        -1,
		}

		if i, j := order[strings.ToLower(fields[i])], order[strings.ToLower(fields[j])]; i != j {
			return i < j
		}

		return strings.ToLower(fields[i]) < strings.ToLower(fields[j])
	})

	for _, fieldName := range fields {
		fmt.Printf(fmt.Sprintf("%%-%dv | ", t.fieldWidths[fieldName]), fieldName)
	}
	fmt.Println()

	for _, row := range t.rows {
		for _, fieldName := range fields {
			val := fmt.Sprintf("%v", row[fieldName])
			val = strings.ReplaceAll(val, "\n", "")
			val = strings.ReplaceAll(val, "\t", "        ")
			fmt.Printf(fmt.Sprintf("%%-%dv | ", t.fieldWidths[fieldName]), val)
		}
		fmt.Println()
	}
	fmt.Println()
}
