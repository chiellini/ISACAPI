package handler

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAIRecordUsageInputsCarryQuotaPlatform(t *testing.T) {
	files := []string{
		"openai_gateway_handler.go",
		"openai_chat_completions.go",
		"openai_embeddings.go",
		"openai_images.go",
	}

	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
			require.NoError(t, err)

			var missing []token.Position
			ast.Inspect(file, func(node ast.Node) bool {
				literal, ok := node.(*ast.CompositeLit)
				if !ok || !isOpenAIRecordUsageInputLiteral(literal.Type) {
					return true
				}
				if !compositeLiteralHasKey(literal, "QuotaPlatform") {
					missing = append(missing, fset.Position(literal.Lbrace))
				}
				return true
			})

			require.Empty(t, missing, "OpenAI usage post-billing must receive request-time QuotaPlatform")
		})
	}
}

func TestCyberPolicyUsageInputCarriesQuotaPlatform(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath.Join(".", "openai_gateway_handler.go"), nil, 0)
	require.NoError(t, err)

	seen := 0
	var missing []token.Position
	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok || !isServiceInputLiteral(literal.Type, "CyberPolicyUsageInput") {
			return true
		}
		seen++
		if !compositeLiteralHasKey(literal, "QuotaPlatform") {
			missing = append(missing, fset.Position(literal.Lbrace))
		}
		return true
	})

	require.Positive(t, seen, "expected at least one CyberPolicyUsageInput literal")
	require.Empty(t, missing, "detached cyber billing must receive request-time QuotaPlatform")
}

func isOpenAIRecordUsageInputLiteral(expr ast.Expr) bool {
	return isServiceInputLiteral(expr, "OpenAIRecordUsageInput")
}

func isServiceInputLiteral(expr ast.Expr, typeName string) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	return ok && pkg.Name == "service" && selector.Sel.Name == typeName
}

func compositeLiteralHasKey(literal *ast.CompositeLit, key string) bool {
	for _, elt := range literal.Elts {
		pair, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		ident, ok := pair.Key.(*ast.Ident)
		if ok && ident.Name == key {
			return true
		}
	}
	return false
}
