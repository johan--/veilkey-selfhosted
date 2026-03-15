package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func normalizeFunctionScope(scope string) (string, error) {
	scope = strings.ToUpper(strings.TrimSpace(scope))
	switch scope {
	case "GLOBAL", "VAULT", "LOCAL", "TEST":
		return scope, nil
	default:
		return "", fmt.Errorf("invalid function scope: %s", scope)
	}
}

func (d *DB) SaveFunction(fn *Function) error {
	if fn == nil {
		return fmt.Errorf("function is required")
	}
	if fn.Name == "" {
		return fmt.Errorf("function name is required")
	}
	if fn.Scope == "" {
		return fmt.Errorf("function scope is required")
	}
	normalizedScope, err := normalizeFunctionScope(fn.Scope)
	if err != nil {
		return err
	}
	fn.Scope = normalizedScope
	if fn.VaultHash == "" {
		return fmt.Errorf("vault_hash is required")
	}
	if fn.FunctionHash == "" {
		return fmt.Errorf("function_hash is required")
	}
	if fn.Command == "" {
		return fmt.Errorf("command is required")
	}
	if fn.VarsJSON == "" {
		fn.VarsJSON = "{}"
	}
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO functions (
			name, scope, vault_hash, function_hash, category, command, vars_json,
			description, tags_json, provenance, last_tested_at, last_run_at, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(name) DO UPDATE SET
			scope = excluded.scope,
			vault_hash = excluded.vault_hash,
			function_hash = excluded.function_hash,
			category = excluded.category,
			command = excluded.command,
			vars_json = excluded.vars_json,
			description = excluded.description,
			tags_json = excluded.tags_json,
			provenance = excluded.provenance,
			last_tested_at = excluded.last_tested_at,
			last_run_at = excluded.last_run_at,
			updated_at = CURRENT_TIMESTAMP
	`, fn.Name, fn.Scope, fn.VaultHash, fn.FunctionHash, fn.Category, fn.Command, fn.VarsJSON,
		fn.Description, coalesceString(fn.TagsJSON, "[]"), coalesceString(fn.Provenance, "local"),
		nullTimeValue(fn.LastTestedAt), nullTimeValue(fn.LastRunAt))
	if err != nil {
		return err
	}

	if err := d.insertFunctionLogTx(tx, fn.FunctionHash, "save", "ok", map[string]string{
		"name":       fn.Name,
		"scope":      fn.Scope,
		"vault_hash": fn.VaultHash,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) GetFunction(name string) (*Function, error) {
	var fn Function
	err := d.conn.QueryRow(`
		SELECT name, scope, vault_hash, function_hash, category, command, vars_json,
		       description, tags_json, provenance, last_tested_at, last_run_at, created_at, updated_at
		FROM functions
		WHERE name = ?
	`, name).Scan(
		&fn.Name,
		&fn.Scope,
		&fn.VaultHash,
		&fn.FunctionHash,
		&fn.Category,
		&fn.Command,
		&fn.VarsJSON,
		&fn.Description,
		&fn.TagsJSON,
		&fn.Provenance,
		&fn.LastTestedAt,
		&fn.LastRunAt,
		&fn.CreatedAt,
		&fn.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("function %s not found", name)
	}
	return &fn, nil
}

func (d *DB) ListFunctions() ([]Function, error) {
	return d.ListFunctionsByScope("")
}

func (d *DB) ListFunctionsByScope(scope string) ([]Function, error) {
	if scope != "" {
		normalizedScope, err := normalizeFunctionScope(scope)
		if err != nil {
			return nil, err
		}
		scope = normalizedScope
	}

	rows, err := d.conn.Query(`
		SELECT name, scope, vault_hash, function_hash, category, command, vars_json,
		       description, tags_json, provenance, last_tested_at, last_run_at, created_at, updated_at
		FROM functions
		WHERE (? = '' OR scope = ?)
		ORDER BY name
	`, scope, scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Function
	for rows.Next() {
		var fn Function
		if err := rows.Scan(
			&fn.Name,
			&fn.Scope,
			&fn.VaultHash,
			&fn.FunctionHash,
			&fn.Category,
			&fn.Command,
			&fn.VarsJSON,
			&fn.Description,
			&fn.TagsJSON,
			&fn.Provenance,
			&fn.LastTestedAt,
			&fn.LastRunAt,
			&fn.CreatedAt,
			&fn.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, fn)
	}
	return out, nil
}

func (d *DB) DeleteFunction(name string) error {
	fn, err := d.GetFunction(name)
	if err != nil {
		return err
	}

	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`DELETE FROM functions WHERE name = ?`, name)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("function %s not found", name)
	}
	if err := d.insertFunctionLogTx(tx, fn.FunctionHash, "delete", "ok", map[string]string{
		"name":       fn.Name,
		"scope":      fn.Scope,
		"vault_hash": fn.VaultHash,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (d *DB) CleanupExpiredTestFunctions(now time.Time) (int, error) {
	rows, err := d.conn.Query(`
		SELECT name, scope, vault_hash, function_hash, category, command, vars_json,
		       description, tags_json, provenance, last_tested_at, last_run_at, created_at, updated_at
		FROM functions
		WHERE scope = 'TEST'
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var expired []Function
	cutoff := now.Add(-1 * time.Hour)
	for rows.Next() {
		var fn Function
		if err := rows.Scan(
			&fn.Name,
			&fn.Scope,
			&fn.VaultHash,
			&fn.FunctionHash,
			&fn.Category,
			&fn.Command,
			&fn.VarsJSON,
			&fn.Description,
			&fn.TagsJSON,
			&fn.Provenance,
			&fn.LastTestedAt,
			&fn.LastRunAt,
			&fn.CreatedAt,
			&fn.UpdatedAt,
		); err != nil {
			return 0, err
		}
		if !fn.CreatedAt.After(cutoff) {
			expired = append(expired, fn)
		}
	}

	if len(expired) == 0 {
		return 0, nil
	}

	tx, err := d.conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	for _, fn := range expired {
		if _, err := tx.Exec(`DELETE FROM functions WHERE name = ?`, fn.Name); err != nil {
			return 0, err
		}
		if err := d.insertFunctionLogTx(tx, fn.FunctionHash, "cleanup", "deleted", map[string]string{
			"name":       fn.Name,
			"scope":      fn.Scope,
			"vault_hash": fn.VaultHash,
		}); err != nil {
			return 0, err
		}
	}

	return len(expired), tx.Commit()
}

func (d *DB) CountFunctions() (int, error) {
	var count int
	err := d.conn.QueryRow(`SELECT COUNT(*) FROM functions`).Scan(&count)
	return count, err
}

func (d *DB) ListFunctionLogs() ([]FunctionLog, error) {
	rows, err := d.conn.Query(`
		SELECT id, function_hash, action, status, detail_json, created_at
		FROM function_logs
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []FunctionLog
	for rows.Next() {
		var entry FunctionLog
		if err := rows.Scan(&entry.ID, &entry.FunctionHash, &entry.Action, &entry.Status, &entry.DetailJSON, &entry.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return out, nil
}

func (d *DB) insertFunctionLogTx(tx *sql.Tx, functionHash, action, status string, detail interface{}) error {
	payload := "{}"
	if detail != nil {
		raw, err := json.Marshal(detail)
		if err != nil {
			return err
		}
		payload = string(raw)
	}
	_, err := tx.Exec(`
		INSERT INTO function_logs (function_hash, action, status, detail_json, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, functionHash, action, status, payload)
	return err
}
