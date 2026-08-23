package storage

import (
	"database/sql"
)

type ScriptRepo struct {
	db *sql.DB
}

func CreateScriptRepo(db *sql.DB) *ScriptRepo {
	return &ScriptRepo{db: db}
}

func (r *ScriptRepo) List() ([]Script, error) {
	rows, err := r.db.Query(`SELECT id, name, command, work_dir FROM scripts ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			//TODO: fix the error
			panic(err)
		}
	}(rows)

	var out []Script
	for rows.Next() {
		var s Script
		if err := rows.Scan(&s.ID, &s.Name, &s.Command, &s.WorkDir); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *ScriptRepo) Get(id string) (*Script, error) {
	row := r.db.QueryRow(`SELECT id, name, command, work_dir FROM scripts WHERE id = ?`, id)
	var s Script
	if err := row.Scan(&s.ID, &s.Name, &s.Command, &s.WorkDir); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ScriptRepo) Upsert(s Script) error {
	_, err := r.db.Exec(`
		INSERT INTO scripts (id, name, command, work_dir)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, command=excluded.command,
			work_dir=excluded.work_dir, 
	`, s.ID, s.Name, s.Command, s.WorkDir)
	return err
}

func (r *ScriptRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM scripts WHERE id = ?`, id)
	return err
}
