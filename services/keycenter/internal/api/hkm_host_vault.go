package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"veilkey-keycenter/internal/crypto"
	"veilkey-keycenter/internal/db"
)

func hostRefRank(ref db.TokenRef) int {
	score := 0
	switch strings.ToUpper(strings.TrimSpace(ref.RefScope)) {
	case "LOCAL":
		score += 300
	case "EXTERNAL":
		score += 200
	case "TEMP":
		score += 100
	}
	switch strings.ToLower(strings.TrimSpace(ref.Status)) {
	case "active":
		score += 30
	case "pending":
		score += 20
	case "temp":
		score += 10
	}
	score += ref.Version
	return score
}

func hostCanonicalRefsByFamily(refs []db.TokenRef, family string) []db.TokenRef {
	byName := map[string]db.TokenRef{}
	for i := range refs {
		ref := refs[i]
		if ref.AgentHash != "" || ref.RefFamily != family {
			continue
		}
		key := strings.TrimSpace(ref.SecretName)
		if key == "" {
			key = strings.TrimSpace(ref.RefID)
		}
		if key == "" {
			continue
		}
		current, exists := byName[key]
		if !exists || hostRefRank(ref) > hostRefRank(current) {
			byName[key] = ref
		}
	}
	items := make([]db.TokenRef, 0, len(byName))
	for _, ref := range byName {
		items = append(items, ref)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].SecretName < items[j].SecretName
	})
	return items
}

func hostRefIndex(refs []db.TokenRef, family string) map[string]db.TokenRef {
	index := map[string]db.TokenRef{}
	for _, ref := range hostCanonicalRefsByFamily(refs, family) {
		index[ref.SecretName] = ref
		if _, exists := index[ref.RefID]; !exists {
			index[ref.RefID] = ref
		}
	}
	return index
}

func hostRefsByFamily(refs []db.TokenRef, family string) []db.TokenRef {
	return hostCanonicalRefsByFamily(refs, family)
}

func secretByNameOrRef(secrets []db.Secret) map[string]db.Secret {
	index := map[string]db.Secret{}
	for i := range secrets {
		secret := secrets[i]
		index[secret.Name] = secret
		if secret.Ref != "" {
			index[secret.Ref] = secret
		}
	}
	return index
}

func hostScopeStatus(ref db.TokenRef, fallbackScope string) (string, string) {
	scope, status, err := normalizeScopeStatus(ref.RefFamily, ref.RefScope, ref.Status, fallbackScope)
	if err != nil {
		return fallbackScope, db.DefaultRefStatusForFamily(ref.RefFamily, fallbackScope)
	}
	return scope, status
}

func (s *Server) upsertHostSecretRecord(name, refID, value string, version int) error {
	info, err := s.db.GetNodeInfo()
	if err != nil {
		return err
	}
	localDEK, err := crypto.DecryptDEK(s.kek, info.DEK, info.DEKNonce)
	if err != nil {
		return err
	}
	ciphertext, nonce, err := crypto.Encrypt(localDEK, []byte(value))
	if err != nil {
		return err
	}
	existing, err := s.db.GetSecretByName(name)
	if err == nil {
		existing.Ref = refID
		existing.Ciphertext = ciphertext
		existing.Nonce = nonce
		existing.Version = version
		return s.db.SaveSecret(existing)
	}
	return s.db.SaveSecret(&db.Secret{
		ID:         crypto.GenerateUUID(),
		Name:       name,
		Ref:        refID,
		Ciphertext: ciphertext,
		Nonce:      nonce,
		Version:    version,
	})
}

func (s *Server) handleHostVaultKeys(w http.ResponseWriter, r *http.Request) {
	secrets, err := s.db.ListSecrets()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to list host vault keys")
		return
	}
	refs, _ := s.db.ListRefs()
	secretIndex := secretByNameOrRef(secrets)
	items := make([]map[string]any, 0)
	for _, ref := range hostRefsByFamily(refs, "VK") {
		scope, status := hostScopeStatus(ref, "TEMP")
		secret := secretIndex[ref.SecretName]
		items = append(items, map[string]any{
			"name":   ref.SecretName,
			"ref":    ref.RefID,
			"token":  ref.RefCanonical,
			"scope":  scope,
			"status": status,
			"value":  "",
			"has_value": secret.Name != "",
		})
	}

	s.respondJSON(w, http.StatusOK, map[string]any{
		"secrets": items,
		"count":   len(items),
	})
}

func (s *Server) handleHostVaultKeySave(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Scope  string `json:"scope"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Value = strings.TrimSpace(req.Value)
	if req.Name == "" || req.Value == "" {
		s.respondError(w, http.StatusBadRequest, "name and value are required")
		return
	}
	if !isValidResourceName(req.Name) {
		s.respondError(w, http.StatusBadRequest, "name must match [A-Z_][A-Z0-9_]*")
		return
	}
	scope, status, err := normalizeScopeStatus("VK", req.Scope, req.Status, "TEMP")
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	refs, _ := s.db.ListRefs()
	refIndex := hostRefIndex(refs, "VK")
	refID := ""
	if ref, ok := refIndex[req.Name]; ok {
		refID = ref.RefID
	} else {
		refID = strings.ToLower(strings.ReplaceAll(crypto.GenerateUUID(), "-", ""))[:8]
	}
	info, err := s.db.GetNodeInfo()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "host node info not available")
		return
	}
	if err := s.upsertHostSecretRecord(req.Name, refID, req.Value, info.Version); err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to save host vault key")
		return
	}
	if err := s.db.SaveRefWithName(db.RefParts{Family: "VK", Scope: scope, ID: refID}, "local-cipher", info.Version, status, "", req.Name); err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to track host vault key")
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]any{
		"name":   req.Name,
		"ref":    refID,
		"token":  "VK:" + scope + ":" + refID,
		"scope":  scope,
		"status": status,
	})
}

func (s *Server) handleHostVaultKeyGet(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("name"))
	if name == "" {
		s.respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	secret, err := s.db.GetSecretByName(name)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "host vault key not found")
		return
	}
	info, err := s.db.GetNodeInfo()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "host node info not available")
		return
	}
	localDEK, err := crypto.DecryptDEK(s.kek, info.DEK, info.DEKNonce)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to decrypt host DEK")
		return
	}
	plaintext, err := crypto.Decrypt(localDEK, secret.Ciphertext, secret.Nonce)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to decrypt host key")
		return
	}

	refs, _ := s.db.ListRefs()
	refIndex := hostRefIndex(refs, "VK")
	scope := "TEMP"
	status := "temp"
	token := ""
	if ref, ok := refIndex[secret.Name]; ok {
		scope, status = hostScopeStatus(ref, "TEMP")
		token = ref.RefCanonical
	} else if ref, ok := refIndex[secret.Ref]; ok {
		scope, status = hostScopeStatus(ref, "TEMP")
		token = ref.RefCanonical
	}
	if token == "" {
		token = "VK:" + scope + ":" + secret.Ref
	}

	s.respondJSON(w, http.StatusOK, map[string]any{
		"name":    secret.Name,
		"ref":     secret.Ref,
		"token":   token,
		"scope":   scope,
		"status":  status,
		"version": secret.Version,
		"value":   string(plaintext),
	})
}

func (s *Server) handleHostVaultConfigs(w http.ResponseWriter, r *http.Request) {
	secrets, _ := s.db.ListSecrets()
	refs, err := s.db.ListRefs()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to list host vault configs")
		return
	}
	secretIndex := secretByNameOrRef(secrets)
	items := make([]map[string]any, 0)
	for _, ref := range hostRefsByFamily(refs, "VE") {
		scope, status := hostScopeStatus(ref, "LOCAL")
		secret := secretIndex[ref.SecretName]
		value := ""
		if secret.Name != "" {
			info, err := s.db.GetNodeInfo()
			if err == nil {
				if localDEK, err := crypto.DecryptDEK(s.kek, info.DEK, info.DEKNonce); err == nil {
					if plaintext, err := crypto.Decrypt(localDEK, secret.Ciphertext, secret.Nonce); err == nil {
						value = string(plaintext)
					}
				}
			}
		}
		items = append(items, map[string]any{
			"key":    ref.SecretName,
			"value":  value,
			"ref":    ref.RefCanonical,
			"scope":  scope,
			"status": status,
		})
	}
	s.respondJSON(w, http.StatusOK, map[string]any{
		"configs": items,
		"count":   len(items),
	})
}

func (s *Server) handleHostVaultConfigGet(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.PathValue("key"))
	if key == "" {
		s.respondError(w, http.StatusBadRequest, "key is required")
	return
	}
	refs, err := s.db.ListRefs()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to load host vault config")
		return
	}
	refIndex := hostRefIndex(refs, "VE")
	ref, ok := refIndex[key]
	if !ok {
		s.respondError(w, http.StatusNotFound, "host vault config not found")
		return
	}
	scope, status := hostScopeStatus(ref, "LOCAL")
	value := ""
	if secret, err := s.db.GetSecretByName(ref.SecretName); err == nil {
		if info, err := s.db.GetNodeInfo(); err == nil {
			if localDEK, err := crypto.DecryptDEK(s.kek, info.DEK, info.DEKNonce); err == nil {
				if plaintext, err := crypto.Decrypt(localDEK, secret.Ciphertext, secret.Nonce); err == nil {
					value = string(plaintext)
				}
			}
		}
	}
	s.respondJSON(w, http.StatusOK, map[string]any{
		"key":    ref.SecretName,
		"value":  value,
		"ref":    ref.RefCanonical,
		"scope":  scope,
		"status": status,
	})
}

func (s *Server) handleHostVaultConfigSave(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key    string `json:"key"`
		Value  string `json:"value"`
		Scope  string `json:"scope"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Key = strings.TrimSpace(req.Key)
	req.Value = strings.TrimSpace(req.Value)
	if req.Key == "" || req.Value == "" {
		s.respondError(w, http.StatusBadRequest, "key and value are required")
		return
	}
	if !isValidResourceName(req.Key) {
		s.respondError(w, http.StatusBadRequest, "key must match [A-Z_][A-Z0-9_]*")
		return
	}
	scope, status, err := normalizeScopeStatus("VE", req.Scope, req.Status, "LOCAL")
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	info, err := s.db.GetNodeInfo()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "host node info not available")
		return
	}
	if err := s.upsertHostSecretRecord(req.Key, req.Key, req.Value, info.Version); err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to save host vault config")
		return
	}
	if err := s.db.SaveRefWithName(db.RefParts{Family: "VE", Scope: scope, ID: req.Key}, "local-config", info.Version, status, "", req.Key); err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to track host vault config")
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]any{
		"key":    req.Key,
		"value":  req.Value,
		"ref":    "VE:" + scope + ":" + req.Key,
		"scope":  scope,
		"status": status,
	})
}
