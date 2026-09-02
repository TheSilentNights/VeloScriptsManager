package storage

import (
	"database/sql"
	"encoding/json"
	"log"

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
		Select("id", "name", "work_dir", "command", "environments").
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
		if err := rows.Close(); err != nil {
			log.Printf("close rows: %v", err)
		}
	}(rows)

	var out = make([]Script, 0)
	for rows.Next() {
		var s Script
		var command string
		var environments string
		if err := rows.Scan(&s.ID, &s.Name, &s.WorkDir, &command, &environments); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(command), &s.Command); err != nil {
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
		Select("id", "name", "work_dir", "command", "environments").
		From("scripts").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := sr.db.QueryRow(query, args...)

	var s Script
	var command string
	var environments string
	if err := row.Scan(&s.ID, &s.Name, &s.WorkDir, &command, &environments); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(command), &s.Command); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(environments), &s.Environments); err != nil {
		return nil, err
	}

	return &s, nil
}

func (sr *ScriptRepo) Insert(s Script) (int64, error) {
	command, err := json.Marshal(s.Command)
	if err != nil {
		return -1, err
	}
	environments, err := json.Marshal(s.Environments)
	if err != nil {
		return -1, err
	}

	query, args, err := squirrel.
		Insert("scripts").
		Columns("id", "name", "work_dir", "command", "environments").
		Values(s.ID, s.Name, s.WorkDir, string(command), string(environments)).
		ToSql()

	if err != nil {
		return -1, err
	}

	result, errExec := sr.db.Exec(query, args...)

	if errExec != nil {
		return -1, errExec
	}

	affected, errEffect := result.RowsAffected()
	if errEffect != nil {
		return -1, errEffect
	}

	return affected, nil
}

func (sr *ScriptRepo) Update(s Script) (int64, error) {
	command, err := json.Marshal(s.Command)
	if err != nil {
		return -1, err
	}
	environments, err := json.Marshal(s.Environments)
	if err != nil {
		return -1, err
	}

	query, args, err := squirrel.
		Update("scripts").
		Set("name", s.Name).
		Set("work_dir", s.WorkDir).
		Set("command", string(command)).
		Set("environments", string(environments)).
		Where(squirrel.Eq{"id": s.ID}).
		ToSql()
	if err != nil {
		return -1, err
	}

	result, err := sr.db.Exec(query, args...)
	if err != nil {
		return -1, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return -1, err
	}

	return affected, nil

}

func (sr *ScriptRepo) Delete(id string) (int64, error) {
	query, args, err := squirrel.
		Delete("scripts").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return -1, err
	}

	result, err := sr.db.Exec(query, args...)
	if err != nil {
		return -1, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return -1, err
	}

	return affected, nil
}
