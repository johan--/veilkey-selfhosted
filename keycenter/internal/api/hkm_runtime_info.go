package api

func (s *Server) hkmRuntimeInfo() (map[string]interface{}, error) {
	info, err := s.db.GetNodeInfo()
	if err != nil {
		return nil, err
	}

	childCount, _ := s.db.CountChildren()
	trackedRefCount, _ := s.db.CountRefs()
	secretCount, _ := s.db.CountSecrets()
	configCount, _ := s.db.CountConfigs()

	resp := map[string]interface{}{
		"mode":               "hkm",
		"node_id":            info.NodeID,
		"vault_node_uuid":    info.NodeID,
		"version":            info.Version,
		"children_count":     childCount,
		"tracked_refs_count": trackedRefCount,
		"secrets_count":      secretCount,
		"configs_count":      configCount,
	}
	if info.ParentURL != "" {
		resp["parent_url"] = info.ParentURL
	}
	return resp, nil
}
