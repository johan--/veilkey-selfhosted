package db

func (d *DB) SetRegistryEntry(entry *KeyRegistryEntry) error {
	return d.conn.Save(entry).Error
}

func (d *DB) GetRegistryEntries(nodeID string) ([]KeyRegistryEntry, error) {
	var entries []KeyRegistryEntry
	err := d.conn.Where("node_id = ?", nodeID).Order("key_name").Find(&entries).Error
	return entries, err
}

func (d *DB) GetRegistryByKeyName(keyName string) ([]KeyRegistryEntry, error) {
	var entries []KeyRegistryEntry
	err := d.conn.Where("key_name = ?", keyName).Order("node_id").Find(&entries).Error
	return entries, err
}

func (d *DB) ListRegistry() ([]KeyRegistryEntry, error) {
	var entries []KeyRegistryEntry
	err := d.conn.Order("node_id, key_name").Find(&entries).Error
	return entries, err
}
