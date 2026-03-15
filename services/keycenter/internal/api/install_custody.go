package api

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"strings"

	vcrypto "veilkey-keycenter/internal/crypto"
	"veilkey-keycenter/internal/db"
)

type installCustodyRequest struct {
	SessionID  string `json:"session_id"`
	Email      string `json:"email"`
	SecretName string `json:"secret_name"`
	BaseURL    string `json:"base_url"`
}

type installCustodySubmitRequest struct {
	Token string `json:"token"`
	Value string `json:"value"`
}

func appendUnique(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func (s *Server) handleCreateInstallCustodyChallenge(w http.ResponseWriter, r *http.Request) {
	var req installCustodyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Email = strings.TrimSpace(req.Email)
	req.SecretName = strings.TrimSpace(req.SecretName)
	if req.SessionID == "" || req.SecretName == "" {
		s.respondError(w, http.StatusBadRequest, "session_id and secret_name are required")
		return
	}
	if _, err := s.db.GetInstallSession(req.SessionID); err != nil {
		s.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	token := vcrypto.GenerateUUID()
	challenge := &db.InstallCustodyChallenge{
		Token:      token,
		SessionID:  req.SessionID,
		Email:      req.Email,
		SecretName: req.SecretName,
		Status:     "pending",
	}
	if err := s.db.SaveInstallCustodyChallenge(challenge); err != nil {
		s.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	baseURL := strings.TrimRight(strings.TrimSpace(req.BaseURL), "/")
	if baseURL == "" {
		baseURL = requestBaseURL(r)
	}
	link := baseURL + "/approve/install/custody?token=" + token
	if req.Email != "" {
		subject := "VeilKey install custody input"
		body := fmt.Sprintf(
			"Open the link below to provide the first-install custody value.\n\nSession: %s\nTarget name: %s\nLink: %s\n",
			req.SessionID,
			req.SecretName,
			link,
		)
		if err := sendInstallMail(req.Email, subject, body); err != nil {
			s.respondError(w, http.StatusBadGateway, err.Error())
			return
		}
	}
	s.respondJSON(w, http.StatusCreated, map[string]any{
		"token": token,
		"link":  link,
	})
}

func (s *Server) handleInstallCustodyPage(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		s.respondError(w, http.StatusBadRequest, "token is required")
		return
	}
	challenge, err := s.db.GetInstallCustodyChallenge(token)
	if err != nil {
		s.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if challenge.Status == "submitted" {
		s.respondError(w, http.StatusGone, "challenge already used")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	emailLabel := challenge.Email
	if strings.TrimSpace(emailLabel) == "" {
		emailLabel = "-"
	}
	fmt.Fprintf(w, installCustodyHTML, challenge.SecretName, emailLabel, token)
}

func (s *Server) handleSubmitInstallCustody(w http.ResponseWriter, r *http.Request) {
	var req installCustodySubmitRequest
	if strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			s.respondError(w, http.StatusBadRequest, "invalid form body")
			return
		}
		req.Token = r.FormValue("token")
		req.Value = r.FormValue("value")
	}
	if req.Token == "" || req.Value == "" {
		s.respondError(w, http.StatusBadRequest, "token and value are required")
		return
	}
	challenge, err := s.db.GetInstallCustodyChallenge(req.Token)
	if err != nil {
		s.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if challenge.Status == "submitted" {
		s.respondError(w, http.StatusGone, "challenge already used")
		return
	}
	key := deriveInstallCustodyKey(s.salt, req.Token)
	ciphertext, nonce, err := vcrypto.Encrypt(key, []byte(req.Value))
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to protect submitted value")
		return
	}
	if _, err := s.db.CompleteInstallCustodyChallenge(req.Token, ciphertext, nonce); err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	session, err := s.db.GetInstallSession(challenge.SessionID)
	if err == nil {
		completed := appendUnique(decodeStringList(session.CompletedStagesJSON), "custody")
		session.CompletedStagesJSON = encodeStringList(completed)
		session.LastStage = "custody"
		_ = s.db.SaveInstallSession(session)
	}
	_ = s.db.SaveAuditEvent(&db.AuditEvent{
		EventID:    vcrypto.GenerateUUID(),
		EntityType: "install_custody",
		EntityID:   req.Token,
		Action:     "submit",
		ActorType:  "user",
		ActorID:    challenge.Email,
		Reason:     "install_custody_submitted",
		Source:     "keycenter_ui",
	})
	if strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		s.respondJSON(w, http.StatusOK, map[string]any{
			"status":     "submitted",
			"session_id": challenge.SessionID,
		})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, installCustodySuccessHTML)
}

func deriveInstallCustodyKey(salt []byte, token string) []byte {
	sum := sha256.Sum256(append(append([]byte{}, salt...), []byte(token)...))
	return sum[:]
}

func requestBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	} else if proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		scheme = proto
	}
	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = r.Host
	}
	return scheme + "://" + host
}

func sendInstallMail(to, subject, body string) error {
	from := strings.TrimSpace(os.Getenv("VEILKEY_OTP_FROM"))
	if from == "" {
		from = "veilkey@localhost"
	}
	if strings.TrimSpace(os.Getenv("VEILKEY_OTP_SMTP_HOST")) != "" {
		return sendInstallSMTP(from, to, subject, body)
	}
	return sendInstallSendmail(from, to, subject, body)
}

func sendInstallSendmail(from, to, subject, body string) error {
	sendmailBin := envDefault("VEILKEY_OTP_SENDMAIL", "/usr/sbin/sendmail")
	if _, err := os.Stat(sendmailBin); err != nil {
		return fmt.Errorf("sendmail binary not found: %s", sendmailBin)
	}
	cmd := exec.Command(sendmailBin, "-t")
	cmd.Stdin = strings.NewReader(formatInstallMail(from, to, subject, body))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sendmail failed: %v: %s", err, string(out))
	}
	return nil
}

func sendInstallSMTP(from, to, subject, body string) error {
	host := strings.TrimSpace(os.Getenv("VEILKEY_OTP_SMTP_HOST"))
	port := envDefault("VEILKEY_OTP_SMTP_PORT", "587")
	username := strings.TrimSpace(os.Getenv("VEILKEY_OTP_SMTP_USERNAME"))
	password := strings.TrimSpace(os.Getenv("VEILKEY_OTP_SMTP_PASSWORD"))
	startTLS := strings.ToLower(envDefault("VEILKEY_OTP_SMTP_STARTTLS", "true")) != "false"
	addr := net.JoinHostPort(host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial failed: %w", err)
	}
	defer conn.Close()
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client failed: %w", err)
	}
	defer client.Close()
	if startTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
				return fmt.Errorf("smtp starttls failed: %w", err)
			}
		}
	}
	if username != "" || password != "" {
		auth := smtp.PlainAuth("", username, password, host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp MAIL FROM failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp RCPT TO failed: %w", err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA failed: %w", err)
	}
	if _, err := io.WriteString(wc, formatInstallMail(from, to, subject, body)); err != nil {
		return fmt.Errorf("smtp write failed: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp close failed: %w", err)
	}
	return client.Quit()
}

func formatInstallMail(from, to, subject, body string) string {
	return fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s", from, to, subject, body)
}

func envDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

const installCustodyHTML = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>VeilKey Install Custody</title>
<style>
body{font-family:-apple-system,system-ui,sans-serif;max-width:720px;margin:40px auto;padding:0 16px;color:#111827}
.card{border:1px solid #d1d5db;border-radius:10px;padding:20px}
label{display:block;margin-top:12px;font-weight:600}
input,textarea,button{width:100%%;padding:10px;margin-top:6px}
button{cursor:pointer}
code{background:#f3f4f6;padding:2px 6px;border-radius:6px}
</style></head><body>
<div class="card">
<h1>VeilKey Install Custody</h1>
<p>Provide the initial custody value for <code>%s</code>.</p>
<p>Recipient: <code>%s</code></p>
<form method="post" action="/approve/install/custody">
<input type="hidden" name="token" value="%s">
<label for="value">Value</label>
<textarea id="value" name="value" rows="6" required></textarea>
<button type="submit">Store custody value</button>
</form>
</div></body></html>
`

const installCustodySuccessHTML = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>VeilKey Install Custody</title></head>
<body style="font-family:-apple-system,system-ui,sans-serif;max-width:680px;margin:40px auto;padding:0 16px">
<h1>Stored</h1>
<p>The install custody value was stored successfully. You can return to the install flow.</p>
</body></html>
`
