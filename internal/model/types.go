package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
)

// StringArray represents a slice of strings stored as JSON or text in the database.
// It implements driver.Valuer and sql.Scanner for seamless GORM / SQL compatibility
// across MySQL, SQLite, and PostgreSQL without requiring external Postgres drivers.
type StringArray []string

// Value implements the driver.Valuer interface.
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface.
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan type into StringArray")
	}

	str := strings.TrimSpace(string(bytes))
	if str == "" {
		*a = []string{}
		return nil
	}

	// Handle legacy PostgreSQL array format "{a,b,c}" if encountered during data migration
	if strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}") {
		trimmed := str[1 : len(str)-1]
		if trimmed == "" {
			*a = []string{}
			return nil
		}
		parts := strings.Split(trimmed, ",")
		res := make([]string, 0, len(parts))
		for _, p := range parts {
			clean := strings.Trim(strings.TrimSpace(p), `"`)
			if clean != "" {
				res = append(res, clean)
			}
		}
		*a = res
		return nil
	}

	return json.Unmarshal([]byte(str), a)
}
