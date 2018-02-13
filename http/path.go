package http

import (
	"fmt"
	"strings"
	"path"
)

// ReadPath returned index and type from location
func (h *Server) readPath(location string) (string, string, error) {
	p := strings.Trim(path.Clean(location), "/")
	parts := strings.Split(p, "/")

	if parts[0] != "logs" && parts[0] != "async-logs" {
		return "", "", fmt.Errorf("%s is not a valid action, only logs and async-logs are allowed", parts[0])
	}

	switch len(parts) {
	case 2:
		return strings.ToLower(parts[1]), DefaultType, nil
	case 3:
		return strings.ToLower(parts[1]), parts[2], nil
	default:
		return "", "", fmt.Errorf("path %s must be in the form /logs/<index>/[type]", location)
	}
}
