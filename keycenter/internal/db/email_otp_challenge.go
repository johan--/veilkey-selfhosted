package db

import (
	"fmt"
	"time"
)

func (d *DB) SaveEmailOTPChallenge(challenge *EmailOTPChallenge) error {
	if challenge == nil {
		return fmt.Errorf("email otp challenge is required")
	}
	if challenge.Token == "" {
		return fmt.Errorf("token is required")
	}
	if challenge.Email == "" {
		return fmt.Errorf("email is required")
	}
	if challenge.Status == "" {
		challenge.Status = "pending"
	}
	return d.conn.Save(challenge).Error
}

func (d *DB) GetEmailOTPChallenge(token string) (*EmailOTPChallenge, error) {
	var challenge EmailOTPChallenge
	if err := d.conn.First(&challenge, "token = ?", token).Error; err != nil {
		return nil, fmt.Errorf("email otp challenge %s not found", token)
	}
	return &challenge, nil
}

func (d *DB) UpdateEmailOTPCode(token, codeHash string, expiresAt time.Time) (*EmailOTPChallenge, error) {
	challenge, err := d.GetEmailOTPChallenge(token)
	if err != nil {
		return nil, err
	}
	challenge.CodeHash = codeHash
	challenge.CodeExpiresAt = &expiresAt
	if err := d.conn.Save(challenge).Error; err != nil {
		return nil, err
	}
	return challenge, nil
}

func (d *DB) MarkEmailOTPVerified(token string) (*EmailOTPChallenge, error) {
	challenge, err := d.GetEmailOTPChallenge(token)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	challenge.Status = "verified"
	challenge.UsedAt = &now
	if err := d.conn.Save(challenge).Error; err != nil {
		return nil, err
	}
	return challenge, nil
}
