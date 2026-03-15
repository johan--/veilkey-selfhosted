package db

import (
	"fmt"
	"time"
)

func (d *DB) SaveApprovalTokenChallenge(challenge *ApprovalTokenChallenge) error {
	if challenge == nil {
		return fmt.Errorf("approval token challenge is required")
	}
	if challenge.Token == "" {
		return fmt.Errorf("token is required")
	}
	if challenge.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if challenge.Status == "" {
		challenge.Status = "pending"
	}
	return d.conn.Save(challenge).Error
}

func (d *DB) GetApprovalTokenChallenge(token string) (*ApprovalTokenChallenge, error) {
	var challenge ApprovalTokenChallenge
	if err := d.conn.First(&challenge, "token = ?", token).Error; err != nil {
		return nil, fmt.Errorf("approval token challenge %s not found", token)
	}
	return &challenge, nil
}

func (d *DB) CompleteApprovalTokenChallenge(token string, ciphertext, nonce []byte) (*ApprovalTokenChallenge, error) {
	challenge, err := d.GetApprovalTokenChallenge(token)
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

func (d *DB) ListApprovalTokenChallenges(targetName, kind, status string, limit, offset int) ([]ApprovalTokenChallenge, int64, error) {
	query := d.conn.Model(&ApprovalTokenChallenge{})
	if targetName != "" {
		query = query.Where("target_name = ?", targetName)
	}
	if kind != "" {
		query = query.Where("kind = ?", kind)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var rows []ApprovalTokenChallenge
	if err := query.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (d *DB) GetLatestApprovalTokenChallenge(targetName, kind, status string) (*ApprovalTokenChallenge, error) {
	rows, _, err := d.ListApprovalTokenChallenges(targetName, kind, status, 1, 0)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("approval token challenge not found")
	}
	return &rows[0], nil
}
