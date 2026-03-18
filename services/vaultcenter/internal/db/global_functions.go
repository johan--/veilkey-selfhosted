package db

import "fmt"

func (d *DB) SaveGlobalFunction(fn *GlobalFunction) error {
	if fn == nil {
		return fmt.Errorf("global function is required")
	}
	if fn.Name == "" {
		return fmt.Errorf("function name is required")
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
	return d.conn.Save(fn).Error
}

func (d *DB) GetGlobalFunction(name string) (*GlobalFunction, error) {
	var fn GlobalFunction
	if err := d.conn.First(&fn, "name = ?", name).Error; err != nil {
		return nil, fmt.Errorf("global function %s not found", name)
	}
	return &fn, nil
}

func (d *DB) ListGlobalFunctions() ([]GlobalFunction, error) {
	var out []GlobalFunction
	if err := d.conn.Order("name ASC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (d *DB) DeleteGlobalFunction(name string) error {
	result := d.conn.Delete(&GlobalFunction{}, "name = ?", name)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("global function %s not found", name)
	}
	return nil
}
