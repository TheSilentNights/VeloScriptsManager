package storage

import (
	"database/sql"

	squirrel "github.com/Masterminds/squirrel"
)

type ScriptRepo struct {
	db *sql.DB
}

func CreateScriptRepo(db *sql.DB) *ScriptRepo {
	return &ScriptRepo{db: db}
}

func (sr *ScriptRepo) List() ([]Script, error) {
	query, args, err := squirrel.
		Select("id", "name", "command", "work_dir", "runner").
		From("scripts").
		OrderBy("name").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := sr.db.Query(query, args...)

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
		if err := rows.Scan(&s.ID, &s.Name, &s.Command, &s.WorkDir, &s.Runner); err != nil {
			return nil, err
		}
		out = append(out, s)
	}

	return out, rows.Err()
}

func (sr *ScriptRepo) Get(id string) (*Script, error) {
	query, args, err := squirrel.
		Select("id", "name", "command", "work_dir", "runner").
		From("scripts").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := sr.db.QueryRow(query, args...)

	var s Script
	if err := row.Scan(&s.ID, &s.Name, &s.Command, &s.WorkDir, &s.Runner); err != nil {
		return nil, err
	}

	return &s, nil
}

func (sr *ScriptRepo) Upsert(s Script) error {
	query, args, err := squirrel.
		Insert("scripts").
		Columns("id", "name", "command", "work_dir", "runner").
		Values(s.ID, s.Name, s.Command, s.WorkDir, s.Runner).
		Suffix("ON CONFLICT(id) DO UPDATE SET name=excluded.name, command=excluded.command, work_dir=excluded.work_dir, runner=excluded.runner").
		ToSql()
	if err != nil {
		return err
	}

	_, err = sr.db.Exec(query, args...)

	return err
}

func (sr *ScriptRepo) Delete(id string) error {
	query, args, err := squirrel.
		Delete("scripts").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = sr.db.Exec(query, args...)

	return err
}
