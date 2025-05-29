package storage

var migrations = []string{
	`CREATE TABLE file (
    path TEXT PRIMARY KEY,
    hash TEXT NOT NULL,
    file_type TEXT,
    updated INTEGER NOT NULL,  -- Unix timestamp
    summary TEXT 
  )`,
}
