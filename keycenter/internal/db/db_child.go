package db

import (
	"fmt"
	"time"
)

func (d *DB) RegisterChild(child *Child) error {
	return d.conn.Save(child).Error
}

func (d *DB) GetChild(nodeID string) (*Child, error) {
	var child Child
	err := d.conn.First(&child, "node_id = ?", nodeID).Error
	if err != nil {
		return nil, fmt.Errorf("child %s not found", nodeID)
	}
	return &child, nil
}

func (d *DB) ListChildren() ([]Child, error) {
	var children []Child
	err := d.conn.Order("created_at").Find(&children).Error
	return children, err
}

func (d *DB) UpdateChildURL(nodeID, url string) error {
	now := time.Now()
	result := d.conn.Model(&Child{}).Where("node_id = ?", nodeID).
		Updates(map[string]interface{}{
			"url":       url,
			"last_seen": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("child %s not found", nodeID)
	}
	return nil
}

func (d *DB) UpdateChildDEK(nodeID string, encryptedDEK, nonce []byte, version int) error {
	result := d.conn.Model(&Child{}).Where("node_id = ?", nodeID).
		Updates(map[string]interface{}{
			"encrypted_dek": encryptedDEK,
			"nonce":         nonce,
			"version":       version,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("child %s not found", nodeID)
	}
	return nil
}

func (d *DB) UpdateChildLastSeen(nodeID string) error {
	now := time.Now()
	return d.conn.Model(&Child{}).Where("node_id = ?", nodeID).
		Update("last_seen", now).Error
}

func (d *DB) DeleteChild(nodeID string) error {
	result := d.conn.Delete(&Child{}, "node_id = ?", nodeID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("child %s not found", nodeID)
	}
	d.conn.Delete(&KeyRegistryEntry{}, "node_id = ?", nodeID)
	return nil
}

func (d *DB) CountChildren() (int, error) {
	var count int64
	err := d.conn.Model(&Child{}).Count(&count).Error
	return int(count), err
}
