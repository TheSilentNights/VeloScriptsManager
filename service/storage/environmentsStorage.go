package storage

import (
	"database/sql"

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
		Select("id", "name", "type", "path").
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
		err := rows.Close()
		if err != nil {
			//TODO: fix the error
			panic(err)
		}
	}(rows)

	var out []Environment
	for rows.Next() {
		var e Environment
		if err := rows.Scan(&e.ID, &e.Name, &e.Type, &e.Path); err != nil {
			return nil, err
		}
		out = append(out, e)
	}

	return out, rows.Err()
}

func (er *EnvironmentRepo) Get(id string) (*Environment, error) {
	query, args, err := squirrel.
		Select("id", "name", "type", "path").
		From("environments").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := er.db.QueryRow(query, args...)

	var e Environment
	if err := row.Scan(&e.ID, &e.Name, &e.Type, &e.Path); err != nil {
		return nil, err
	}

	return &e, nil
}

func (er *EnvironmentRepo) Upsert(e Environment) error {
	query, args, err := squirrel.
		Insert("environments").
		Columns("id", "name", "type", "path").
		Values(e.ID, e.Name, e.Type, e.Path).
		Suffix("ON CONFLICT(id) DO UPDATE SET name=excluded.name, type=excluded.type, path=excluded.path").
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

	_, err = er.db.Exec(query, args...)

	return err
}