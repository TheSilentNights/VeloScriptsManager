package storage

import (
	"database/sql"
	"encoding/json"

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
		Select("id", "name", "work_dir", "runner", "params", "environments").
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
		var params string
		var environments string
		if err := rows.Scan(&s.ID, &s.Name, &s.WorkDir, &s.Runner, &params, &environments); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(params), &s.Params); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(environments), &s.Environments); err != nil {
			return nil, err
		}
		out = append(out, s)
	}

	return out, rows.Err()
}

func (sr *ScriptRepo) Get(id string) (*Script, error) {
	query, args, err := squirrel.
		Select("id", "name", "work_dir", "runner", "params", "environments").
		From("scripts").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := sr.db.QueryRow(query, args...)

	var s Script
	var params string
	var environments string
	if err := row.Scan(&s.ID, &s.Name, &s.WorkDir, &s.Runner, &params, &environments); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(params), &s.Params); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(environments), &s.Environments); err != nil {
		return nil, err
	}

	return &s, nil
}

func (sr *ScriptRepo) Upsert(s Script) error {
	params, err := json.Marshal(s.Params)
	if err != nil {
		return err
	}
	environments, err := json.Marshal(s.Environments)
	if err != nil {
		return err
	}

	query, args, err := squirrel.
		Insert("scripts").
		Columns("id", "name", "work_dir", "runner", "params", "environments").
		Values(s.ID, s.Name, s.WorkDir, s.Runner, string(params), string(environments)).
		Suffix("ON CONFLICT(id) DO UPDATE SET name=excluded.name, work_dir=excluded.work_dir, runner=excluded.runner, params=excluded.params, environments=excluded.environments").
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
