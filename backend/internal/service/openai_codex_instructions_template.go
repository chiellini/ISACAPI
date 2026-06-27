package service

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type forcedCodexInstructionsTemplateData struct {
	ExistingInstructions string
	OriginalModel        string
	NormalizedModel      string
	BillingModel         string
	UpstreamModel        string
}

const defaultCodexSecurityBoundaryInstructions = `
{{- if .ExistingInstructions }}
{{ .ExistingInstructions }}

{{- end }}
You are helping with authorized enterprise engineering and hardware-related security work, including legitimate ESP and hardware/software co-design context.
Default to defensive, lawful, and compliance-oriented support:
- Assist with authorized debugging, hardening, threat modeling, and incident response.
- Prioritize prevention, monitoring, baseline configuration cleanup, and policy-compliant operations.
- For suspicious or clearly harmful intent, refuse the harmful action and offer safer alternatives.
When uncertainty exists, request explicit authorization and clarify legal/compliance constraints first.
`

func applyForcedCodexInstructionsTemplate(
	reqBody map[string]any,
	templateText string,
	data forcedCodexInstructionsTemplateData,
) (bool, error) {
	rendered, err := renderForcedCodexInstructionsTemplate(templateText, data)
	if err != nil {
		return false, err
	}
	if rendered == "" {
		return false, nil
	}

	existing, _ := reqBody["instructions"].(string)
	if strings.TrimSpace(existing) == rendered {
		return false, nil
	}

	reqBody["instructions"] = rendered
	return true, nil
}

func renderForcedCodexInstructionsTemplate(
	templateText string,
	data forcedCodexInstructionsTemplateData,
) (string, error) {
	tmpl, err := template.New("forced_codex_instructions").Option("missingkey=zero").Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("parse forced codex instructions template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render forced codex instructions template: %w", err)
	}

	return strings.TrimSpace(buf.String()), nil
}

func resolveCodexInstructionsTemplateText(templateText string, injectSafeDefault bool) string {
	trimmedTemplate := strings.TrimSpace(templateText)
	if trimmedTemplate != "" {
		return trimmedTemplate
	}
	if !injectSafeDefault {
		return ""
	}
	return strings.TrimSpace(defaultCodexSecurityBoundaryInstructions)
}

func buildCodexInstructionsFromTemplate(
	existingInstructions string,
	templateText string,
	injectSafeDefault bool,
	data forcedCodexInstructionsTemplateData,
) (string, error) {
	resolvedTemplate := resolveCodexInstructionsTemplateText(templateText, injectSafeDefault)
	if strings.TrimSpace(resolvedTemplate) == "" {
		return strings.TrimSpace(existingInstructions), nil
	}
	if strings.TrimSpace(data.ExistingInstructions) == "" {
		data.ExistingInstructions = strings.TrimSpace(existingInstructions)
	}
	reqBody := map[string]any{"instructions": existingInstructions}
	updated, err := applyForcedCodexInstructionsTemplate(reqBody, resolvedTemplate, data)
	if err != nil {
		return "", err
	}
	if !updated {
		return strings.TrimSpace(data.ExistingInstructions), nil
	}
	value, _ := reqBody["instructions"].(string)
	return strings.TrimSpace(value), nil
}
