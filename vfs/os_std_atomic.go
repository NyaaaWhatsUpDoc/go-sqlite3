//go:build !linux || !(amd64 || arm64 || riscv64 || ppc64le) || sqlite3_nosys

package vfs

import "os"

func osBatchAtomic(*os.File) bool {
	return false
}
