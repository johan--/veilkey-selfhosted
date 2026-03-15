package api

import "net/http"

// otpExemptOperations lists operations that do NOT require OTP/email
// verification on the client side. Clients should GET /api/otp-policy and
// skip their local OTP gate for any operation whose reason string appears
// in the "exempt" list.
var otpExemptOperations = []string{
	"secret reveal",
	"secret field reveal",
	"lifecycle activate",
	"lifecycle archive",
	"direct resolve",
}

func (s *Server) handleOTPPolicy(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"exempt": otpExemptOperations,
	})
}
