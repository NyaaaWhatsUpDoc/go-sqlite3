package csv

import (
	"strings"

	"github.com/ncruces/go-sqlite3/util/sql3util"
)

type affinity byte

const (
	blob    affinity = 0
	text    affinity = 1
	numeric affinity = 2
	integer affinity = 3
	real    affinity = 4
)

func getColumnAffinities(schema string) ([]affinity, error) {
	tab, err := sql3util.ParseTable(schema)
	if err != nil {
		return nil, err
	}

	columns := tab.Columns
	types := make([]affinity, len(columns))
	for i, col := range columns {
		types[i] = getAffinity(col.Type)
	}
	return types, nil
}

func getAffinity(declType string) affinity {
	// https://sqlite.org/datatype3.html#determination_of_column_affinity
	if declType == "" {
		return blob
	}
	name := strings.ToUpper(declType)
	if strings.Contains(name, "INT") {
		return integer
	}
	if strings.Contains(name, "CHAR") || strings.Contains(name, "CLOB") || strings.Contains(name, "TEXT") {
		return text
	}
	if strings.Contains(name, "BLOB") {
		return blob
	}
	if strings.Contains(name, "REAL") || strings.Contains(name, "FLOA") || strings.Contains(name, "DOUB") {
		return real
	}
	return numeric
}
