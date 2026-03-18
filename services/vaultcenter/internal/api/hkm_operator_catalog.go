package api

import (
	"net/http"
	"strconv"
	"veilkey-vaultcenter/internal/db"
)

func secretCatalogPayload(entry db.SecretCatalog) map[string]any {
	return map[string]any{
		"secret_canonical_id": entry.SecretCanonicalID,
		"secret_name":         entry.SecretName,
		"display_name":        entry.DisplayName,
		"description":         entry.Description,
		"tags_json":           entry.TagsJSON,
		"class":               entry.Class,
		"scope":               entry.Scope,
		"status":              entry.Status,
		"vault_node_uuid":     entry.VaultNodeUUID,
		"vault_runtime_hash":  entry.VaultRuntimeHash,
		"vault_hash":          entry.VaultHash,
		"ref_canonical":       entry.RefCanonical,
		"fields_present_json": entry.FieldsPresentJSON,
		"binding_count":       entry.BindingCount,
		"usage_count":         entry.BindingCount,
		"last_rotated_at":     entry.LastRotatedAt,
		"last_revealed_at":    entry.LastRevealedAt,
		"updated_at":          entry.UpdatedAt,
	}
}

func (s *Server) handleVaultInventory(w http.ResponseWriter, r *http.Request) {
	limit, offset, errMsg := parseListWindow(r)
	if errMsg != "" {
		s.respondError(w, http.StatusBadRequest, errMsg)
		return
	}
	rows, total, err := s.db.ListVaultInventoryFiltered(
		r.URL.Query().Get("status"),
		r.URL.Query().Get("vault_hash"),
		limit,
		offset,
	)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to list vault inventory")
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]any{
		"vaults":      rows,
		"count":       len(rows),
		"total_count": total,
		"limit":       limit,
		"offset":      offset,
	})
}

func (s *Server) handleSecretCatalogList(w http.ResponseWriter, r *http.Request) {
	limit, offset, errMsg := parseListWindow(r)
	if errMsg != "" {
		s.respondError(w, http.StatusBadRequest, errMsg)
		return
	}
	rows, total, err := s.db.ListSecretCatalogFiltered(
		r.URL.Query().Get("vault_hash"),
		r.URL.Query().Get("class"),
		r.URL.Query().Get("status"),
		r.URL.Query().Get("q"),
		limit,
		offset,
	)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to list secret catalog")
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, secretCatalogPayload(row))
	}
	s.respondJSON(w, http.StatusOK, map[string]any{
		"secrets":     items,
		"count":       len(items),
		"total_count": total,
		"limit":       limit,
		"offset":      offset,
	})
}

func (s *Server) handleSecretCatalogGet(w http.ResponseWriter, r *http.Request) {
	refCanonical := r.PathValue("ref")
	if refCanonical == "" {
		s.respondError(w, http.StatusBadRequest, "ref is required")
		return
	}

	entry, err := s.db.GetSecretCatalogByRef(refCanonical)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "secret catalog entry not found")
		return
	}
	s.respondJSON(w, http.StatusOK, secretCatalogPayload(*entry))
}

func (s *Server) handleBindingsList(w http.ResponseWriter, r *http.Request) {
	bindingType := r.URL.Query().Get("binding_type")
	targetName := r.URL.Query().Get("target_name")
	refCanonical := r.URL.Query().Get("ref_canonical")

	limit, offset, errMsg := parseListWindow(r)
	if errMsg != "" {
		s.respondError(w, http.StatusBadRequest, errMsg)
		return
	}
	var (
		rows  []db.Binding
		total int64
		err   error
	)
	switch {
	case refCanonical != "":
		rows, total, err = s.db.ListBindingsByRefFiltered(
			refCanonical,
			r.URL.Query().Get("vault_hash"),
			limit,
			offset,
		)
	case bindingType != "" && targetName != "":
		rows, total, err = s.db.ListBindingsFiltered(
			bindingType,
			targetName,
			r.URL.Query().Get("vault_hash"),
			refCanonical,
			limit,
			offset,
		)
	default:
		s.respondError(w, http.StatusBadRequest, "either ref_canonical or binding_type and target_name are required")
		return
	}
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to list bindings")
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]any{
		"bindings":    rows,
		"count":       len(rows),
		"total_count": total,
		"limit":       limit,
		"offset":      offset,
	})
}

func (s *Server) handleAuditEventsList(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")
	if entityType == "" || entityID == "" {
		s.respondError(w, http.StatusBadRequest, "entity_type and entity_id are required")
		return
	}

	limit, offset, errMsg := parseListWindow(r)
	if errMsg != "" {
		s.respondError(w, http.StatusBadRequest, errMsg)
		return
	}
	rows, total, err := s.db.ListAuditEventsLimited(entityType, entityID, limit, offset)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to list audit events")
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]any{
		"events":      rows,
		"count":       len(rows),
		"total_count": total,
		"limit":       limit,
		"offset":      offset,
	})
}

func parseListWindow(r *http.Request) (int, int, string) {
	limit := 100
	offset := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			return 0, 0, "limit must be a non-negative integer"
		}
		limit = parsed
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			return 0, 0, "offset must be a non-negative integer"
		}
		offset = parsed
	}
	return limit, offset, ""
}
