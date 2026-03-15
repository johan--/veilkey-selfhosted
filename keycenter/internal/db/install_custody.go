package db

import (
	"fmt"
	"time"
)

func (d *DB) SaveInstallCustodyChallenge(challenge *InstallCustodyChallenge) error {
	if challenge == nil {
		return fmt.Errorf("install custody challenge is required")
	}
	if challenge.Token == "" {
		return fmt.Errorf("token is required")
	}
	if challenge.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	if challenge.SecretName == "" {
		return fmt.Errorf("secret_name is required")
	}
	if challenge.Status == "" {
		challenge.Status = "pending"
	}
	return d.conn.Save(challenge).Error
}

func (d *DB) GetInstallCustodyChallenge(token string) (*InstallCustodyChallenge, error) {
	var challenge InstallCustodyChallenge
	if err := d.conn.First(&challenge, "token = ?", token).Error; err != nil {
		return nil, fmt.Errorf("install custody challenge %s not found", token)
	}
	return &challenge, nil
}

func (d *DB) CompleteInstallCustodyChallenge(token string, ciphertext, nonce []byte) (*InstallCustodyChallenge, error) {
	challenge, err := d.GetInstallCustodyChallenge(token)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	challenge.Ciphertext = ciphertext
	challenge.Nonce = nonce
	challenge.Status = "submitted"
	challenge.UsedAt = &now
	if err := d.conn.Save(challenge).Error; err != nil {
		return nil, err
	}
	return challenge, nil
}
