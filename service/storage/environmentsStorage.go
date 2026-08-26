package storage

import (
	"database/sql"
	"encoding/json"
	"log"

	squirrel "github.com/Masterminds/squirrel"
)

type EnvironmentRepo struct {
	db *sql.DB
}

func CreateEnvironmentRepo(db *sql.DB) *EnvironmentRepo {
	return &EnvironmentRepo{db: db}
}

func (er *EnvironmentRepo) List() ([]Environment, error) {
	query, args, err := squirrel.
		Select("id", "name", "type", "path", "env", "children").
		From("environments").
		OrderBy("name").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := er.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			log.Printf("close rows: %v", err)
		}
	}(rows)

	var out = make([]Environment, 0)
	for rows.Next() {
		e, err := scanEnvironment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *e)
	}

	return out, rows.Err()
}

func (er *EnvironmentRepo) Get(id string) (*Environment, error) {
	query, args, err := squirrel.
		Select("id", "name", "type", "path", "env", "children").
		From("environments").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := er.db.QueryRow(query, args...)
	return scanEnvironment(row)
}

func (er *EnvironmentRepo) Upsert(e Environment) error {
	env, err := json.Marshal(e.Env)
	if err != nil {
		return err
	}
	children, err := json.Marshal(e.Children)
	if err != nil {
		return err
	}

	query, args, err := squirrel.
		Insert("environments").
		Columns("id", "name", "type", "path", "env", "children").
		Values(e.ID, e.Name, e.Type, e.Path, string(env), string(children)).
		Suffix("ON CONFLICT(id) DO UPDATE SET name=excluded.name, type=excluded.type, path=excluded.path, env=excluded.env, children=excluded.children").
		ToSql()
	if err != nil {
		return err
	}

	_, err = er.db.Exec(query, args...)

	return err
}

func (er *EnvironmentRepo) Delete(id string) error {
	query, args, err := squirrel.
		Delete("environments").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := er.db.Exec(query, args...)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

// scanner abstracts *sql.Row and *sql.Rows so a single decode helper works for
// both Get and List.
type scanner interface {
	Scan(dest ...any) error
}

func scanEnvironment(s scanner) (*Environment, error) {
	var e Environment
	var env string
	var children string
	if err := s.Scan(&e.ID, &e.Name, &e.Type, &e.Path, &env, &children); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(env), &e.Env); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(children), &e.Children); err != nil {
		return nil, err
	}
	return &e, nil
}
