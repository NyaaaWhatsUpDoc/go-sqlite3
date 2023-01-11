#include "sqlite3.h"

int main(int argc, char *argv[]) {
	int rc = sqlite3_initialize();
	if (rc != SQLITE_OK) return -1;
}

sqlite3_vfs *sqlite3_demovfs();

int sqlite3_os_init() {
	return sqlite3_vfs_register(sqlite3_demovfs(), /*default=*/1);
}
