package db

import (
	"fmt"
	"time"
)

func (d *DB) SaveSecretInputChallenge(challenge *SecretInputChallenge) error {
	if challenge == nil {
		return fmt.Errorf("secret input challenge is required")
	}
	if challenge.Token == "" {
		return fmt.Errorf("token is required")
	}
	if challenge.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	if challenge.Vault == "" {
		return fmt.Errorf("vault is required")
	}
	if challenge.SecretName == "" {
		return fmt.Errorf("secret_name is required")
	}
	if challenge.Status == "" {
		challenge.Status = "pending"
	}
	return d.conn.Save(challenge).Error
}

func (d *DB) GetSecretInputChallenge(token string) (*SecretInputChallenge, error) {
	var challenge SecretInputChallenge
	if err := d.conn.First(&challenge, "token = ?", token).Error; err != nil {
		return nil, fmt.Errorf("secret input challenge %s not found", token)
	}
	return &challenge, nil
}

func (d *DB) CompleteSecretInputChallenge(token string) (*SecretInputChallenge, error) {
	challenge, err := d.GetSecretInputChallenge(token)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	challenge.Status = "submitted"
	challenge.UsedAt = &now
	if err := d.conn.Save(challenge).Error; err != nil {
		return nil, err
	}
	return challenge, nil
}
