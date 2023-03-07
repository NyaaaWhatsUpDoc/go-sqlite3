package sqlite3

import (
	"context"
	"database/sql/driver"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"sync/atomic"
	"unsafe"
)

// Conn is a database connection handle.
// A Conn is not safe for concurrent use by multiple goroutines.
//
// https://www.sqlite.org/c3ref/sqlite3.html
type Conn struct {
	*module

	handle    uint32
	arena     arena
	interrupt context.Context
	waiter    chan struct{}
	pending   *Stmt
}

// Open calls [OpenFlags] with [OPEN_READWRITE], [OPEN_CREATE] and [OPEN_URI].
func Open(filename string) (*Conn, error) {
	return newConn(filename, OPEN_READWRITE|OPEN_CREATE|OPEN_URI)
}

// OpenFlags opens an SQLite database file as specified by the filename argument.
//
// If a URI filename is used, PRAGMA statements to execute can be specified using "_pragma":
//
//	sqlite3.Open("file:demo.db?_pragma=busy_timeout(10000)&_pragma=locking_mode(normal)")
//
// https://www.sqlite.org/c3ref/open.html
func OpenFlags(filename string, flags OpenFlag) (*Conn, error) {
	return newConn(filename, flags)
}

func newConn(filename string, flags OpenFlag) (conn *Conn, err error) {
	mod, err := instantiateModule()
	if err != nil {
		return nil, err
	}
	defer func() {
		if conn == nil {
			mod.close()
		} else {
			runtime.SetFinalizer(conn, finalizer[Conn](3))
		}
	}()

	c := &Conn{module: mod}
	c.arena = c.newArena(1024)
	c.handle, err = c.openDB(filename, flags)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Conn) openDB(filename string, flags OpenFlag) (uint32, error) {
	defer c.arena.reset()
	connPtr := c.arena.new(ptrlen)
	namePtr := c.arena.string(filename)

	r := c.call(c.api.open, uint64(namePtr), uint64(connPtr), uint64(flags), 0)

	handle := c.mem.readUint32(connPtr)
	if err := c.module.error(r[0], handle); err != nil {
		c.closeDB(handle)
		return 0, err
	}

	if flags|OPEN_URI != 0 && strings.HasPrefix(filename, "file:") {
		var pragmas strings.Builder
		if _, after, ok := strings.Cut(filename, "?"); ok {
			query, _ := url.ParseQuery(after)
			for _, p := range query["_pragma"] {
				pragmas.WriteString(`PRAGMA `)
				pragmas.WriteString(p)
				pragmas.WriteByte(';')
			}
		}

		pragmaPtr := c.arena.string(pragmas.String())
		r := c.call(c.api.exec, uint64(handle), uint64(pragmaPtr), 0, 0, 0)
		if err := c.module.error(r[0], handle, pragmas.String()); err != nil {
			c.closeDB(handle)
			return 0, fmt.Errorf("sqlite3: invalid _pragma: %w", err)
		}
	}
	return handle, nil
}

func (c *Conn) closeDB(handle uint32) {
	r := c.call(c.api.closeZombie, uint64(handle))
	if err := c.module.error(r[0], handle); err != nil {
		panic(err)
	}
}

// Close closes the database connection.
//
// If the database connection is associated with unfinalized prepared statements,
// open blob handles, and/or unfinished backup objects,
// Close will leave the database connection open and return [BUSY].
//
// It is safe to close a nil, zero or closed Conn.
//
// https://www.sqlite.org/c3ref/close.html
func (c *Conn) Close() error {
	if c == nil || c.handle == 0 {
		return nil
	}

	c.SetInterrupt(context.Background())

	r := c.call(c.api.close, uint64(c.handle))
	if err := c.error(r[0]); err != nil {
		return err
	}

	c.handle = 0
	runtime.SetFinalizer(c, nil)
	return c.module.close()
}

// Exec is a convenience function that allows an application to run
// multiple statements of SQL without having to use a lot of code.
//
// https://www.sqlite.org/c3ref/exec.html
func (c *Conn) Exec(sql string) error {
	c.checkInterrupt()
	defer c.arena.reset()
	sqlPtr := c.arena.string(sql)

	r := c.call(c.api.exec, uint64(c.handle), uint64(sqlPtr), 0, 0, 0)
	return c.error(r[0])
}

// MustPrepare calls [Conn.Prepare] and panics on error,
// a nil Stmt, or a non-empty tail.
func (c *Conn) MustPrepare(sql string) *Stmt {
	s, tail, err := c.PrepareFlags(sql, 0)
	if err != nil {
		panic(err)
	}
	if s == nil {
		panic(emptyErr)
	}
	if !emptyStatement(tail) {
		s.Close()
		panic(tailErr)
	}
	return s
}

// Prepare calls [Conn.PrepareFlags] with no flags.
func (c *Conn) Prepare(sql string) (stmt *Stmt, tail string, err error) {
	return c.PrepareFlags(sql, 0)
}

// PrepareFlags compiles the first SQL statement in sql;
// tail is left pointing to what remains uncompiled.
// If the input text contains no SQL (if the input is an empty string or a comment),
// both stmt and err will be nil.
//
// https://www.sqlite.org/c3ref/prepare.html
func (c *Conn) PrepareFlags(sql string, flags PrepareFlag) (stmt *Stmt, tail string, err error) {
	if emptyStatement(sql) {
		return nil, "", nil
	}

	defer c.arena.reset()
	stmtPtr := c.arena.new(ptrlen)
	tailPtr := c.arena.new(ptrlen)
	sqlPtr := c.arena.string(sql)

	r := c.call(c.api.prepare, uint64(c.handle),
		uint64(sqlPtr), uint64(len(sql)+1), uint64(flags),
		uint64(stmtPtr), uint64(tailPtr))

	stmt = &Stmt{c: c}
	stmt.handle = c.mem.readUint32(stmtPtr)
	i := c.mem.readUint32(tailPtr)
	tail = sql[i-sqlPtr:]

	if err := c.error(r[0], sql); err != nil {
		return nil, "", err
	}
	if stmt.handle == 0 {
		return nil, "", nil
	}
	return
}

// GetAutocommit tests the connection for auto-commit mode.
//
// https://www.sqlite.org/c3ref/get_autocommit.html
func (c *Conn) GetAutocommit() bool {
	r := c.call(c.api.autocommit, uint64(c.handle))
	return r[0] != 0
}

// LastInsertRowID returns the rowid of the most recent successful INSERT
// on the database connection.
//
// https://www.sqlite.org/c3ref/last_insert_rowid.html
func (c *Conn) LastInsertRowID() int64 {
	r := c.call(c.api.lastRowid, uint64(c.handle))
	return int64(r[0])
}

// Changes returns the number of rows modified, inserted or deleted
// by the most recently completed INSERT, UPDATE or DELETE statement
// on the database connection.
//
// https://www.sqlite.org/c3ref/changes.html
func (c *Conn) Changes() int64 {
	r := c.call(c.api.changes, uint64(c.handle))
	return int64(r[0])
}

// SetInterrupt interrupts a long-running query when a context is done.
//
// Subsequent uses of the connection will return [INTERRUPT]
// until the context is reset by another call to SetInterrupt.
//
// To associate a timeout with a connection:
//
//	ctx, cancel := context.WithTimeout(context.TODO(), 100*time.Millisecond)
//	conn.SetInterrupt(ctx)
//	defer cancel()
//
// SetInterrupt returns the old context assigned to the connection.
//
// https://www.sqlite.org/c3ref/interrupt.html
func (c *Conn) SetInterrupt(ctx context.Context) (old context.Context) {
	// Is a waiter running?
	if c.waiter != nil {
		c.waiter <- struct{}{} // Cancel the waiter.
		<-c.waiter             // Wait for it to finish.
		c.waiter = nil
	}

	old = c.interrupt
	c.interrupt = ctx
	if ctx == nil || ctx.Done() == nil {
		// Finalize the uncompleted SQL statement.
		if c.pending != nil {
			c.pending.Close()
			c.pending = nil
		}
		return old
	}

	// Creating an uncompleted SQL statement prevents SQLite from ignoring
	// an interrupt that comes before any other statements are started.
	if c.pending == nil {
		c.pending = c.MustPrepare(`SELECT 1 UNION ALL SELECT 2`)
		c.pending.Step()
	} else {
		c.pending.Reset()
	}

	// Don't create the goroutine if we're already interrupted.
	// This happens frequently while restoring to a previously interrupted state.
	if c.checkInterrupt() {
		return old
	}

	waiter := make(chan struct{})
	c.waiter = waiter
	go func() {
		select {
		case <-waiter: // Waiter was cancelled.
			break

		case <-ctx.Done(): // Done was closed.
			buf := c.mem.view(c.handle+c.api.interrupt, 4)
			(*atomic.Uint32)(unsafe.Pointer(&buf[0])).Store(1)
			// Wait for the next call to SetInterrupt.
			<-waiter
		}

		// Signal that the waiter has finished.
		waiter <- struct{}{}
	}()
	return old
}

func (c *Conn) checkInterrupt() bool {
	if c.interrupt == nil || c.interrupt.Err() == nil {
		return false
	}
	buf := c.mem.view(c.handle+c.api.interrupt, 4)
	(*atomic.Uint32)(unsafe.Pointer(&buf[0])).Store(1)
	return true
}

// Pragma executes a PRAGMA statement and returns any results.
//
// https://www.sqlite.org/pragma.html
func (c *Conn) Pragma(str string) []string {
	stmt := c.MustPrepare(`PRAGMA ` + str)
	defer stmt.Close()

	var pragmas []string
	for stmt.Step() {
		pragmas = append(pragmas, stmt.ColumnText(0))
	}
	return pragmas
}

func (c *Conn) error(rc uint64, sql ...string) error {
	return c.module.error(rc, c.handle, sql...)
}

// DriverConn is implemented by the SQLite [database/sql] driver connection.
//
// It can be used to access advanced SQLite features like
// [savepoints] and [incremental BLOB I/O].
//
// [savepoints]: https://www.sqlite.org/lang_savepoint.html
// [incremental BLOB I/O]: https://www.sqlite.org/c3ref/blob_open.html
type DriverConn interface {
	driver.ConnBeginTx
	driver.ExecerContext
	driver.ConnPrepareContext

	Savepoint() (release func(*error))
	OpenBlob(db, table, column string, row int64, write bool) (*Blob, error)
}
