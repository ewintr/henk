package storage

var migrations = []string{
	`CREATE TABLE files (
    id TEXT PRIMARY KEY,
    path TEXT UNIQUE NOT NULL,
    checksum TEXT NOT NULL,
    file_type TEXT,
    last_updated INTEGER NOT NULL,  -- Unix timestamp
    summary TEXT 
  )`,
}
