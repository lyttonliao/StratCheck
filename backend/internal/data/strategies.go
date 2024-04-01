package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/lyttonliao/StratCheck/internal/validator"
)

// Use '-' directive to hide internal system information that isn't relevant or sensitive info
// Use 'omitempty' directive to hide a field in the JSON output if and only if struct field value is
// empty, where empty is defined as equal to
// false, 0, or "" | empty array, slice or map | nil pointer or nil interface value
// Use the 'string' directive to force data to be represented as a string in JSON output
// string only works on struct fields which have int*, uint*, float* or bool types
type Strategy struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Name      string    `json:"name"`
	Fields    []string  `json:"fields,omitempty"`
	Criteria  []string  `json:"criteria,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateStrategy(v *validator.Validator, strategy *Strategy) {
	v.Check(strategy.Name != "", "name", "must be provided")
	v.Check(len(strategy.Name) <= 500, "name", "must not be more than 500 bytes long")
	v.Check(len(strategy.Fields) >= 1, "fields", "must contain at least 1 field")
	v.Check(len(strategy.Criteria) >= 1, "criteria", "must contain at least 1 criterium")
	v.Check(validator.Unique(strategy.Fields), "fields", "must not contain duplicate values")
	v.Check(validator.Unique(strategy.Criteria), "criteria", "must not contain duplicate values")
}

type StrategyModel struct {
	DB *sql.DB
}

func (s StrategyModel) Insert(strategy *Strategy) error {
	query := `
		INSERT INTO strategies (name, fields, criteria)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, version
	`

	// pq.Array() takes our []string slice and converts it to a pq.StringArray type
	// which implements the driver.Valuer and sql.Scanner interfaces which provide values
	// the PostgreSQL database can understand and store in a text[] array column
	args := []interface{}{
		strategy.Name,
		pq.Array(strategy.Fields),
		pq.Array(strategy.Criteria),
	}

	return s.DB.QueryRow(query, args...).Scan(&strategy.ID, &strategy.CreatedAt, &strategy.Version)
}

func (s StrategyModel) Get(id int64) (*Strategy, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, created_at, fields, criteria, version
		FROM strategies
		WHERE id = $1
	`

	var strategy Strategy

	err := s.DB.QueryRow(query, id).Scan(
		&strategy.ID,
		&strategy.Name,
		&strategy.CreatedAt,
		pq.Array(&strategy.Fields),
		pq.Array(&strategy.Criteria),
		&strategy.Version,
	)
	// If no match, Scan() will return a sql.ErrNoRows error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &strategy, nil
}

func (s StrategyModel) Update(strategy *Strategy) error {
	query := `
		UPDATE strategies
		SET name = $1, fields = $2, criteria = $3, version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING version
	`

	args := []interface{}{
		strategy.Name,
		pq.Array(strategy.Fields),
		pq.Array(strategy.Criteria),
		strategy.ID,
		strategy.Version,
	}

	err := s.DB.QueryRow(query, args...).Scan(&strategy.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (s StrategyModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE from strategies
		WHERE id = $1
	`

	// Exec() method executes the query, passing in args for
	// placeholder parameters, returns a sql.Result object
	result, err := s.DB.Exec(query, id)
	if err != nil {
		return err
	}

	//
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
