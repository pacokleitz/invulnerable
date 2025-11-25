package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type SBOM struct {
	ID        int             `db:"id" json:"id"`
	ScanID    int             `db:"scan_id" json:"scan_id"`
	Format    string          `db:"format" json:"format"` // cyclonedx or spdx
	Version   *string         `db:"version" json:"version,omitempty"`
	Document  json.RawMessage `db:"document" json:"document"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
}

// JSONB is a custom type for PostgreSQL JSONB handling
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONB value")
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = result
	return nil
}
