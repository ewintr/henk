package storage

import (
	"database/sql"
	"fmt"
	"time"

	"go-mod.ewintr.nl/henk/internal"
)

type SqliteFile struct {
	db *sql.DB
}

func NewSqliteFile(db *sql.DB) *SqliteFile {
	return &SqliteFile{
		db: db,
	}
}

func (r *SqliteFile) Store(file internal.File) error {
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

func (r *SqliteFile) FindByPath(path string) (internal.File, error) {
	row := r.db.QueryRow(`
SELECT path, hash, file_type, updated, summary
FROM file
WHERE path = ?
`, path)

	var file internal.File
	var lastUpdatedUnix int64
	err := row.Scan(&file.Path, &file.Hash, &file.FileType, &lastUpdatedUnix, &file.Summary, path)
	switch {
	case err == sql.ErrNoRows:
		return internal.File{}, ErrNotFound
	case err != nil:
		return internal.File{}, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}

	file.Updated = time.Unix(lastUpdatedUnix, 0)

	return file, nil
}

func (r *SqliteFile) FindAll() ([]internal.File, error) {
	rows, err := r.db.Query(`
SELECT path, hash, file_type, updated, summary
FROM file
ORDER BY path ASC
`)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSqliteFailure, err)
	}
	defer rows.Close()

	var files []internal.File
	for rows.Next() {
		var file internal.File
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

	return files, nil
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
