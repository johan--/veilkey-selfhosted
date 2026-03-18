package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

func (d *DB) SaveSecret(secret *Secret) error {
	return d.conn.Save(secret).Error
}

func (d *DB) GetSecretByName(name string) (*Secret, error) {
	var s Secret
	err := d.conn.First(&s, "name = ?", name).Error
	if err != nil {
		return nil, fmt.Errorf("secret %s not found", name)
	}
	return &s, nil
}

func (d *DB) GetSecretByID(id string) (*Secret, error) {
	var s Secret
	err := d.conn.First(&s, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("secret id %s not found", id)
	}
	return &s, nil
}

func (d *DB) GetSecretByRef(refHash string) (*Secret, error) {
	var s Secret
	err := d.conn.First(&s, "ref = ?", refHash).Error
	if err != nil {
		return nil, fmt.Errorf("secret ref %s not found", refHash)
	}
	return &s, nil
}

func (d *DB) ListSecrets() ([]Secret, error) {
	var secrets []Secret
	err := d.conn.Order("name").Find(&secrets).Error
	return secrets, err
}

func (d *DB) DeleteSecret(name string) error {
	result := d.conn.Delete(&Secret{}, "name = ?", name)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("secret %s not found", name)
	}
	return nil
}

func (d *DB) CountSecrets() (int, error) {
	var count int64
	err := d.conn.Model(&Secret{}).Count(&count).Error
	return int(count), err
}

// ReencryptAllSecrets re-encrypts all secrets with a new DEK.
func (d *DB) ReencryptAllSecrets(
	decryptFn func(ciphertext, nonce []byte) ([]byte, error),
	encryptFn func(plaintext []byte) (ciphertext, nonce []byte, err error),
	newVersion int,
) (int, error) {
	secrets, err := d.ListSecrets()
	if err != nil {
		return 0, err
	}

	count := 0
	err = d.conn.Transaction(func(tx *gorm.DB) error {
		for i := range secrets {
			s := &secrets[i]
			plaintext, err := decryptFn(s.Ciphertext, s.Nonce)
			if err != nil {
				return fmt.Errorf("decrypt secret %s: %w", s.Name, err)
			}
			newCiphertext, newNonce, err := encryptFn(plaintext)
			if err != nil {
				return fmt.Errorf("encrypt secret %s: %w", s.Name, err)
			}
			if err := tx.Model(s).Updates(map[string]interface{}{
				"ciphertext": newCiphertext,
				"nonce":      newNonce,
				"version":    newVersion,
				"updated_at": time.Now(),
			}).Error; err != nil {
				return err
			}
			count++
		}
		return nil
	})
	return count, err
}
