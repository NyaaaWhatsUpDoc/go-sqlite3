package bcw2

import (
	"path/filepath"
	"testing"

	"github.com/ncruces/go-sqlite3/driver"
	"github.com/ncruces/go-sqlite3/vfs"
)

func Test_bcw2(t *testing.T) {
	if !vfs.SupportsSharedMemory {
		t.Skip("skipping without shared memory")
	}

	tmp := filepath.ToSlash(filepath.Join(t.TempDir(), "test.db"))

	db, err := driver.Open("file:" + tmp + "?_pragma=journal_mode(wal2)&_txlock=concurrent")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`CREATE TABLE test (col)`)
	if err != nil {
		t.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}

	var version string
	err = db.QueryRow(`SELECT sqlite_version()`).Scan(&version)
	if err != nil {
		t.Fatal(err)
	}
	if version != "3.48.0" {
		t.Error(version)
	}
}
