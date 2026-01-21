package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// cryptoRandRead exists to make ID generation error branches unit-testable.
// Production code uses crypto/rand.Read.
var cryptoRandRead = rand.Read

func resolveOrGenerateResourceID(diags *diag.Diagnostics, configVal types.String, attrPath path.Path) (string, bool) {
	if configVal.IsUnknown() {
		diags.AddAttributeError(attrPath, "Unknown value", "resource_id must be known during planning when set explicitly.")
		return "", false
	}
	if configVal.IsNull() {
		id, err := generateTerraformResourceID("tf-")
		if err != nil {
			diags.AddAttributeError(attrPath, "Failed to generate resource_id", err.Error())
			return "", false
		}
		return id, true
	}
	return configVal.ValueString(), true
}

func generateTerraformResourceID(prefix string) (string, error) {
	b := make([]byte, 12)
	if _, err := cryptoRandRead(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s", prefix, hex.EncodeToString(b)), nil
}
