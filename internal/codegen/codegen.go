// Package codegen generates application-layer code from resolved pgdesign schemas.
// It extracts RLS policies and produces language-specific validators that can
// pre-check policy conditions before hitting the database.
package codegen

import (
	"strings"

	"github.com/smm-h/pgdesign/internal/model"
)

// Generator generates application code from a resolved schema.
type Generator interface {
	// Generate produces source code for all eligible policies in the schema.
	Generate(schema *model.Schema) ([]byte, error)
}

// PolicyContext holds the data needed to generate a validator for one policy.
type PolicyContext struct {
	SchemaName   string
	TableName    string
	PolicyName   string
	Operation    string
	Using        string
	WithCheck    string
	ErrorCode    string
	ErrorMessage string
}

// ExtractPolicies collects all policies from a schema into PolicyContexts.
func ExtractPolicies(schema *model.Schema) []PolicyContext {
	var contexts []PolicyContext
	for _, tbl := range schema.Tables {
		for _, pol := range tbl.Policies {
			contexts = append(contexts, PolicyContext{
				SchemaName:   tbl.Schema,
				TableName:    tbl.Name,
				PolicyName:   pol.Name,
				Operation:    pol.Operation,
				Using:        pol.Using,
				WithCheck:    pol.WithCheck,
				ErrorCode:    pol.ErrorCode,
				ErrorMessage: pol.ErrorMessage,
			})
		}
	}
	return contexts
}

// FilterGeneratable returns only policies that have an ErrorCode and reference
// player_privacy_settings in their Using or WithCheck expression. These are the
// social enforcement policies that benefit from pre-check validators. Policies
// that are purely row-ownership checks (e.g. "player_id = current_setting(...)") or
// system/read-all policies (using = "true") are skipped.
func FilterGeneratable(policies []PolicyContext) []PolicyContext {
	var result []PolicyContext
	for _, p := range policies {
		if p.ErrorCode == "" {
			continue
		}
		expr := p.Using + " " + p.WithCheck
		if !strings.Contains(expr, "player_privacy_settings") {
			continue
		}
		result = append(result, p)
	}
	return result
}
