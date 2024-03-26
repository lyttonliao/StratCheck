package data

import (
	"time"
)

// Use '-' directive to hide internal system information that isn't relevant or sensitive info
// Use 'omitempty' directive to hide a field in the JSON output if and only if struct field value is
// empty, where empty is defined as equal to 
// false, 0, or "" | empty array, slice or map | nil pointer or nil interface value
// Use the 'string' directive to force data to be represented as a string in JSON output
// string only works on struct fields which have int*, uint*, float* or bool types
type Strategy struct {
	ID 			int64		`json:"id"`
	CreatedAt 	time.Time	`json:"-"`
	Name		string		`json:"name"`
	Fields		[]string	`json:"fields,omitempty"`
	Criteria	[]string	`json:"criteria,omitempty"`
	Version		int32		`json:"version"`
}