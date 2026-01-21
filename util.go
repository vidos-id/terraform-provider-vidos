package main

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func joinURL(base, path string) string {
	return strings.TrimRight(base, "/") + path
}

func joinURLWithQuery(base, urlPath string, query map[string]string) string {
	u, err := url.Parse(joinURL(base, urlPath))
	if err != nil {
		return joinURL(base, urlPath)
	}
	q := u.Query()
	for k, v := range query {
		if strings.TrimSpace(v) == "" {
			continue
		}
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func parseJSONToAny(diags *diag.Diagnostics, raw string, attrPath path.Path, name string) any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		diags.AddAttributeError(attrPath, "Invalid JSON", name+" must be valid JSON")
		return nil
	}
	var out any
	if err := json.Unmarshal([]byte(trimmed), &out); err != nil {
		diags.AddAttributeError(attrPath, "Invalid JSON", name+" must be valid JSON: "+err.Error())
		return nil
	}
	return out
}
