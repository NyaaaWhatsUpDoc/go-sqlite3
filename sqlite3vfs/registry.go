package sqlite3vfs

import "sync"

var (
	// +checklocks:vfsRegistryMtx
	vfsRegistry    map[string]VFS
	vfsRegistryMtx sync.Mutex
)

// Find returns a VFS given its name.
// If there is no match, nil is returned.
//
// https://www.sqlite.org/c3ref/vfs_find.html
func Find(name string) VFS {
	vfsRegistryMtx.Lock()
	defer vfsRegistryMtx.Unlock()
	return vfsRegistry[name]
}

// Register registers a VFS.
//
// https://www.sqlite.org/c3ref/vfs_find.html
func Register(name string, vfs VFS) {
	vfsRegistryMtx.Lock()
	defer vfsRegistryMtx.Unlock()
	if vfsRegistry == nil {
		vfsRegistry = map[string]VFS{}
	}
	vfsRegistry[name] = vfs
}

// Unregister unregisters a VFS.
//
// https://www.sqlite.org/c3ref/vfs_find.html
func Unregister(name string) {
	vfsRegistryMtx.Lock()
	defer vfsRegistryMtx.Unlock()
	delete(vfsRegistry, name)
}