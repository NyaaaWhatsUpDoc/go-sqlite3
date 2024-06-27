package vtabutil

import (
	"context"
	"sync"

	_ "embed"

	"github.com/ncruces/go-sqlite3/internal/util"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

const (
	_NONE = iota
	_MEMORY
	_SYNTAX
	_UNSUPPORTEDSQL

	codeptr = 4
	baseptr = 8
)

var (
	//go:embed parse/sql3parse_table.wasm
	binary  []byte
	ctx     context.Context
	once    sync.Once
	runtime wazero.Runtime
	module  wazero.CompiledModule
)

// Table holds metadata about a table.
type Table struct {
	mod api.Module
	sql string
	fns util.Funcs
	ptr uint32
}

// Parse parses a [CREATE] or [ALTER TABLE] command.
//
// [CREATE]: https://sqlite.org/lang_createtable.html
// [ALTER TABLE]: https://sqlite.org/lang_altertable.html
func Parse(sql string) (_ *Table, err error) {
	once.Do(func() {
		ctx = context.Background()
		cfg := wazero.NewRuntimeConfigInterpreter().WithDebugInfoEnabled(false)
		runtime = wazero.NewRuntimeWithConfig(ctx, cfg)
		module, err = runtime.CompileModule(ctx, binary)
	})
	if err != nil {
		return nil, err
	}

	mod, err := runtime.InstantiateModule(ctx, module, wazero.NewModuleConfig().WithName(""))
	if err != nil {
		return nil, err
	}

	if buf, ok := mod.Memory().Read(baseptr, uint32(len(sql))); ok {
		copy(buf, sql)
	}

	tab := Table{mod: mod, sql: sql}
	r := tab.fns.Call(ctx, mod, "sql3parse_table", baseptr, uint64(len(sql)), codeptr)
	tab.ptr = uint32(r)

	c, _ := mod.Memory().ReadUint32Le(codeptr)
	switch c {
	case _MEMORY:
		panic(util.OOMErr)
	case _SYNTAX:
		return nil, util.ErrorString("sql3parse: invalid syntax")
	case _UNSUPPORTEDSQL:
		return nil, util.ErrorString("sql3parse: unsupported SQL")
	}
	return &tab, nil
}

// Close closes a table handle.
func (t *Table) Close() error {
	mod := t.mod
	t.mod = nil
	return mod.Close(ctx)
}

// NumColumns returns the number of columns of the table.
func (t *Table) NumColumns() int {
	r := t.fns.Call(ctx, t.mod, "sql3table_num_columns", uint64(t.ptr))
	return int(int32(r))
}

// Column returns data for the ith column of the table.
//
// https://sqlite.org/lang_createtable.html#column_definitions
func (t *Table) Column(i int) Column {
	r := t.fns.Call(ctx, t.mod, "sql3table_get_column", uint64(t.ptr), uint64(i))
	return Column{
		tab: t,
		ptr: uint32(r),
	}
}

func (t *Table) string(ptr uint32) string {
	if ptr == 0 {
		return ""
	}
	off, _ := t.mod.Memory().ReadUint32Le(ptr + 0)
	len, _ := t.mod.Memory().ReadUint32Le(ptr + 4)
	return t.sql[off-baseptr : off+len-baseptr]
}

// Column holds metadata about a column.
type Column struct {
	tab *Table
	ptr uint32
}

// Type returns the declared type of a column.
//
// https://sqlite.org/lang_createtable.html#column_data_types
func (c Column) Type() string {
	r := c.tab.fns.Call(ctx, c.tab.mod, "sql3column_type", uint64(c.ptr))
	return c.tab.string(uint32(r))
}
