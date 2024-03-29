package data

import (
	"database/sql"
	"time"

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
	return nil
}

func (s StrategyModel) Get(id int64) (*Strategy, error) {
	return nil, nil
}

func (s StrategyModel) Update(strategy *Strategy) error {
	return nil
}

func (s StrategyModel) Delete(id int64) error {
	return nil
}
