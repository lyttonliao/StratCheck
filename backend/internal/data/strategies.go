package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	Public    bool      `json:"public"`
	Fields    []string  `json:"fields,omitempty"`
	Criteria  []string  `json:"criteria,omitempty"`
	UserID    int64     `json:"user_id"`
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

func IsOwner(userID int64, strategy *Strategy) bool {
	return userID == strategy.UserID
}

type StrategyModel struct {
	DB *sql.DB
}

func (s StrategyModel) Insert(userID int64, strategy *Strategy) error {
	query := `
		INSERT INTO strategies (name, fields, criteria, public, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, version
	`

	args := []interface{}{
		strategy.Name,
		pq.Array(strategy.Fields),
		pq.Array(strategy.Criteria),
		strategy.Public,
		userID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.DB.QueryRowContext(ctx, query, args...).Scan(&strategy.ID, &strategy.CreatedAt, &strategy.Version)
}

func (s StrategyModel) GetAll(userID int64, name string, fields []string, filters Filters) ([]*Strategy, Metadata, error) {
	// to_tsvector('simple', s) takes a string and splits it into lexemes, which is a basic lexical unit of words
	// planto_tsquery('simple', s) takes a string and converts it to a formatted query term by
	// stripping special characters and inserts the & operator between words
	// @@ operator is the matching operator, checks if the query terms match the lexemes
	// @> operator is the contains operator
	query := fmt.Sprintf(
		`SELECT count(*) OVER(), id, created_at, name, fields, criteria, public, version
		FROM strategies
		WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (fields @> $2 OR fields = '{}') AND (public = true OR user_id = $3)
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{name, pq.Array(fields), userID, filters.limit(), filters.offset()}

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	strategies := []*Strategy{}

	for rows.Next() {
		var strategy Strategy

		err := rows.Scan(
			&totalRecords,
			&strategy.ID,
			&strategy.CreatedAt,
			&strategy.Name,
			pq.Array(&strategy.Fields),
			pq.Array(&strategy.Criteria),
			&strategy.Public,
			&strategy.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		strategies = append(strategies, &strategy)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return strategies, metadata, nil
}

func (s StrategyModel) Get(userID int64, strategyID int64) (*Strategy, error) {
	if strategyID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, created_at, public, fields, criteria, user_id, version
		FROM strategies
		WHERE id = $1 AND user_id = $2
	`

	var strategy Strategy

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.DB.QueryRowContext(ctx, query, strategyID, userID).Scan(
		&strategy.ID,
		&strategy.Name,
		&strategy.CreatedAt,
		&strategy.Public,
		pq.Array(&strategy.Fields),
		pq.Array(&strategy.Criteria),
		&strategy.UserID,
		&strategy.Version,
	)

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

func (s StrategyModel) Update(userID int64, strategy *Strategy) error {
	query := `
		UPDATE strategies
		SET name = $1, public = $2, fields = $3, criteria = $4, version = version + 1
		WHERE id = $5 AND user_id = $6 AND version = $7
		RETURNING version
	`

	args := []interface{}{
		strategy.Name,
		strategy.Public,
		pq.Array(strategy.Fields),
		pq.Array(strategy.Criteria),
		strategy.ID,
		userID,
		strategy.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.DB.QueryRowContext(ctx, query, args...).Scan(&strategy.Version)
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

func (s StrategyModel) Delete(userID int64, strategyID int64) error {
	if strategyID < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE from strategies
		WHERE id = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := s.DB.ExecContext(ctx, query, strategyID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
