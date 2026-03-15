package db

import "fmt"

func (d *DB) SaveInstallSession(session *InstallSession) error {
	if session == nil {
		return fmt.Errorf("install session is required")
	}
	if session.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	if session.Version == 0 {
		session.Version = 1
	}
	if session.Language == "" {
		session.Language = "ko"
	}
	if session.Flow == "" {
		session.Flow = "quickstart"
	}
	if session.DeploymentMode == "" {
		session.DeploymentMode = "host-service"
	}
	if session.InstallScope == "" {
		session.InstallScope = "host-only"
	}
	if session.BootstrapMode == "" {
		session.BootstrapMode = "email"
	}
	if session.MailTransport == "" {
		session.MailTransport = "none"
	}
	if session.PlannedStagesJSON == "" {
		session.PlannedStagesJSON = "[]"
	}
	if session.CompletedStagesJSON == "" {
		session.CompletedStagesJSON = "[]"
	}
	return d.conn.Save(session).Error
}

func (d *DB) GetInstallSession(sessionID string) (*InstallSession, error) {
	var session InstallSession
	if err := d.conn.First(&session, "session_id = ?", sessionID).Error; err != nil {
		return nil, fmt.Errorf("install session %s not found", sessionID)
	}
	return &session, nil
}

func (d *DB) GetLatestInstallSession() (*InstallSession, error) {
	var session InstallSession
	if err := d.conn.Order("updated_at DESC").First(&session).Error; err != nil {
		return nil, fmt.Errorf("no install session found")
	}
	return &session, nil
}
