package http

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

func (h *Server) authorize(header string, esIndex string) (int, error) {
	//  headers => { "Authorization" => "ELK devops-uepinruJhq82BAPWnjaw89sJ" }
	parts := strings.Split(header, " ")

	if len(parts) != 2 {
		return http.StatusUnauthorized, fmt.Errorf("bad authorization header (must be in the form ELK project-apikey)")
	}
	if parts[0] != "ELK" {
		return http.StatusUnauthorized, fmt.Errorf("bad authorization header (must start with ELK)")
	}

	key := parts[1]
	if _, keyExists := h.Config.APIKeys[key]; !keyExists {
		return http.StatusUnauthorized, fmt.Errorf("apikey '%s' not found in configuration", key)
	}

	for _, indexPattern := range h.Config.APIKeys[key] {
		if matched, _ := filepath.Match(indexPattern, esIndex); matched {
			return 0, nil
		}
	}
	return http.StatusForbidden, fmt.Errorf("index '%s' is not allowed for apikey '%s'", esIndex, key)
}
