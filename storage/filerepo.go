package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type SqliteFile struct {
	db *sql.DB
}

func NewSqliteFile(db *sql.DB) *SqliteFile {
	return &SqliteFile{
		db: db,
	}
}

func (r *SqliteFile) Store(file File) error {
	if file.Updated.IsZero() {
		file.Updated = time.Now()
	}
	_, err := r.db.Exec(`
        INSERT OR REPLACE INTO file (path, hash, file_type, updated, summary)
        VALUES (?, ?, ?, ?, ?)
    `, file.Path, file.Hash, file.FileType, file.Updated.Unix(), file.Summary)

	if err != nil {
		return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	return nil
}

func (r *SqliteFile) FindByPath(path string) (File, error) {
	row := r.db.QueryRow(`
SELECT path, hash, file_type, updated, summary
FROM file
WHERE path = ?
`, path)

	var file File
	var lastUpdatedUnix int64
	err := row.Scan(&file.Path, &file.Hash, &file.FileType, &lastUpdatedUnix, &file.Summary, path)
	switch {
	case err == sql.ErrNoRows:
		return File{}, ErrNotFound
	case err != nil:
		return File{}, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	file.Updated = time.Unix(lastUpdatedUnix, 0)

	return file, nil
}

func (r *SqliteFile) FindAll() (map[string]File, error) {
	rows, err := r.db.Query(`
SELECT path, hash, file_type, updated, summary
FROM file
ORDER BY path ASC
`)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var file File
		var lastUpdatedUnix int64
		if err := rows.Scan(&file.Path, &file.Hash, &file.FileType, &lastUpdatedUnix, &file.Summary); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
		}
		file.Updated = time.Unix(lastUpdatedUnix, 0)
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	res := make(map[string]File, len(files))
	for _, f := range files {
		res[f.Path] = f
	}

	return res, nil
}

func (r *SqliteFile) ListPaths() ([]string, error) {
	rows, err := r.db.Query(`
SELECT path
FROM file
ORDER BY path ASC
`)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
		}
		paths = append(paths, path)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	return paths, nil
}

func (r *SqliteFile) Delete(path string) error {
	result, err := r.db.Exec(`DELETE FROM file WHERE path = ?`, path)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
