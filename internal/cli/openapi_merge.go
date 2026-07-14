package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pb33f/libopenapi"
	openapivalidator "github.com/pb33f/libopenapi-validator"
)

const (
	openAPIVersion    = "3.1.1"
	jsonSchemaDialect = "https://json-schema.org/draft/2020-12/schema"
)

var componentFieldNames = map[string]struct{}{
	"schemas": {}, "responses": {}, "parameters": {}, "examples": {},
	"requestBodies": {}, "headers": {}, "securitySchemes": {}, "links": {},
	"callbacks": {}, "pathItems": {},
}

type openAPIMergeState struct {
	fields        map[string]json.RawMessage
	components    map[string]map[string]json.RawMessage
	componentExts map[string]json.RawMessage
	servers       []json.RawMessage
	serverKeys    map[string]struct{}
	tags          map[string]json.RawMessage
}

func newOpenAPIMergeState(title, version string) (*openAPIMergeState, error) {
	info, err := json.Marshal(map[string]any{"title": title, "version": version})
	if err != nil {
		return nil, fmt.Errorf("marshal generated info: %w", err)
	}
	openapi, _ := json.Marshal(openAPIVersion)
	dialect, _ := json.Marshal(jsonSchemaDialect)
	return &openAPIMergeState{
		fields: map[string]json.RawMessage{
			"openapi":           openapi,
			"info":              info,
			"jsonSchemaDialect": dialect,
		},
		components:    map[string]map[string]json.RawMessage{},
		componentExts: map[string]json.RawMessage{},
		serverKeys:    map[string]struct{}{},
		tags:          map[string]json.RawMessage{},
	}, nil
}

func (m *openAPIMergeState) mergeFile(path string) error {
	raw, err := os.ReadFile(path) // #nosec G304 -- CLI intentionally reads caller-selected fragment files
	if err != nil {
		return err
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return fmt.Errorf("%s: expected a JSON object (OpenAPI fragment), got %s", path, firstToken(trimmed))
	}
	var fragment map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fragment); err != nil {
		return fmt.Errorf("%s: invalid JSON: %w", path, err)
	}
	for field, value := range fragment {
		var mergeErr error
		switch field {
		case "openapi", "info", "jsonSchemaDialect", "security", "externalDocs":
			mergeErr = m.mergeIdentical(field, value)
		case "servers":
			mergeErr = m.mergeServers(value)
		case "paths", "webhooks":
			mergeErr = m.mergeNamedObject(field, value)
		case "components":
			mergeErr = m.mergeComponents(value)
		case "tags":
			mergeErr = m.mergeTags(value)
		default:
			if strings.HasPrefix(field, "x-") {
				mergeErr = m.mergeIdentical(field, value)
			} else {
				mergeErr = fmt.Errorf("unsupported OpenAPI 3.1 top-level field %q", field)
			}
		}
		if mergeErr != nil {
			return fmt.Errorf("%s: %w", path, mergeErr)
		}
	}
	return nil
}

func (m *openAPIMergeState) mergeIdentical(field string, value json.RawMessage) error {
	if existing, exists := m.fields[field]; exists {
		if !jsonEqual(existing, value) {
			return fmt.Errorf("conflicting %s declarations", field)
		}
		return nil
	}
	m.fields[field] = cloneRaw(value)
	return nil
}

func (m *openAPIMergeState) mergeServers(value json.RawMessage) error {
	var servers []json.RawMessage
	if err := json.Unmarshal(value, &servers); err != nil {
		return fmt.Errorf("servers: expected an array: %w", err)
	}
	for _, server := range servers {
		key, err := canonicalJSON(server)
		if err != nil {
			return fmt.Errorf("servers: invalid entry: %w", err)
		}
		if _, duplicate := m.serverKeys[key]; duplicate {
			continue
		}
		m.serverKeys[key] = struct{}{}
		m.servers = append(m.servers, cloneRaw(server))
	}
	return nil
}

func (m *openAPIMergeState) mergeNamedObject(field string, value json.RawMessage) error {
	var incoming map[string]json.RawMessage
	if err := json.Unmarshal(value, &incoming); err != nil {
		return fmt.Errorf("%s: expected an object: %w", field, err)
	}
	existing := map[string]json.RawMessage{}
	if raw, ok := m.fields[field]; ok {
		if err := json.Unmarshal(raw, &existing); err != nil {
			return fmt.Errorf("%s: invalid accumulated object: %w", field, err)
		}
	}
	for name, item := range incoming {
		if _, duplicate := existing[name]; duplicate {
			return fmt.Errorf("duplicate %s.%s", field, name)
		}
		existing[name] = cloneRaw(item)
	}
	merged, err := json.Marshal(existing)
	if err != nil {
		return fmt.Errorf("%s: marshal merged object: %w", field, err)
	}
	m.fields[field] = merged
	return nil
}

func (m *openAPIMergeState) mergeComponents(value json.RawMessage) error {
	var incoming map[string]json.RawMessage
	if err := json.Unmarshal(value, &incoming); err != nil {
		return fmt.Errorf("components: expected an object: %w", err)
	}
	for field, rawEntries := range incoming {
		if strings.HasPrefix(field, "x-") {
			if existing, duplicate := m.componentExts[field]; duplicate && !jsonEqual(existing, rawEntries) {
				return fmt.Errorf("conflicting components.%s declarations", field)
			}
			m.componentExts[field] = cloneRaw(rawEntries)
			continue
		}
		if _, supported := componentFieldNames[field]; !supported {
			return fmt.Errorf("unsupported components.%s field", field)
		}
		var entries map[string]json.RawMessage
		if err := json.Unmarshal(rawEntries, &entries); err != nil {
			return fmt.Errorf("components.%s: expected an object: %w", field, err)
		}
		if m.components[field] == nil {
			m.components[field] = map[string]json.RawMessage{}
		}
		for name, entry := range entries {
			if _, duplicate := m.components[field][name]; duplicate {
				return fmt.Errorf("duplicate components.%s.%s", field, name)
			}
			m.components[field][name] = cloneRaw(entry)
		}
	}
	return nil
}

func (m *openAPIMergeState) mergeTags(value json.RawMessage) error {
	var tags []json.RawMessage
	if err := json.Unmarshal(value, &tags); err != nil {
		return fmt.Errorf("tags: expected an array: %w", err)
	}
	for _, rawTag := range tags {
		var tag struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(rawTag, &tag); err != nil || tag.Name == "" {
			return fmt.Errorf("tags: every tag must be an object with a non-empty name")
		}
		if existing, duplicate := m.tags[tag.Name]; duplicate {
			if !jsonEqual(existing, rawTag) {
				return fmt.Errorf("conflicting tags.%s declarations", tag.Name)
			}
			continue
		}
		m.tags[tag.Name] = cloneRaw(rawTag)
	}
	return nil
}

func (m *openAPIMergeState) document() ([]byte, error) {
	capacity, err := openAPIDocumentCapacity(len(m.fields))
	if err != nil {
		return nil, err
	}
	doc := make(map[string]any, capacity)
	for field, raw := range m.fields {
		doc[field] = raw
	}
	if _, exists := doc["paths"]; !exists {
		doc["paths"] = map[string]any{}
	}
	if len(m.servers) > 0 {
		doc["servers"] = m.servers
	}
	if len(m.tags) > 0 {
		names := make([]string, 0, len(m.tags))
		for name := range m.tags {
			names = append(names, name)
		}
		sort.Strings(names)
		tags := make([]json.RawMessage, 0, len(names))
		for _, name := range names {
			tags = append(tags, m.tags[name])
		}
		doc["tags"] = tags
	}
	if len(m.components) > 0 || len(m.componentExts) > 0 {
		components := make(map[string]any, len(m.components)+len(m.componentExts))
		for field, entries := range m.components {
			components[field] = entries
		}
		for field, value := range m.componentExts {
			components[field] = value
		}
		doc["components"] = components
	}
	encoded, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	encoded = append(encoded, '\n')
	if err := validateOpenAPI31(encoded); err != nil {
		return nil, err
	}
	return encoded, nil
}

func openAPIDocumentCapacity(fieldCount int) (int, error) {
	const syntheticFields = 3
	maxInt := int(^uint(0) >> 1)
	if fieldCount < 0 || fieldCount > maxInt-syntheticFields {
		return 0, fmt.Errorf("OpenAPI document has too many top-level fields")
	}
	return fieldCount + syntheticFields, nil
}

func validateOpenAPI31(raw []byte) error {
	var header struct {
		OpenAPI           string `json:"openapi"`
		JSONSchemaDialect string `json:"jsonSchemaDialect"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return fmt.Errorf("OpenAPI 3.1.1 validation: invalid JSON: %w", err)
	}
	if header.OpenAPI != openAPIVersion {
		return fmt.Errorf("OpenAPI 3.1.1 validation: openapi must equal %q, got %q", openAPIVersion, header.OpenAPI)
	}
	if header.JSONSchemaDialect != jsonSchemaDialect {
		return fmt.Errorf("OpenAPI 3.1.1 validation: jsonSchemaDialect must equal %q, got %q", jsonSchemaDialect, header.JSONSchemaDialect)
	}
	document, err := libopenapi.NewDocument(raw)
	if err != nil {
		return fmt.Errorf("OpenAPI 3.1.1 validation: %w", err)
	}
	validator, buildErrors := openapivalidator.NewValidator(document)
	if len(buildErrors) > 0 {
		return fmt.Errorf("OpenAPI 3.1.1 validation: build validator: %v", buildErrors)
	}
	defer validator.Release()
	valid, failures := validator.ValidateDocument()
	if !valid {
		messages := make([]string, 0, len(failures))
		for _, failure := range failures {
			message := failure.Message
			if failure.SpecPath != "" {
				message = failure.SpecPath + ": " + message
			}
			messages = append(messages, message)
		}
		return fmt.Errorf("OpenAPI 3.1.1 validation: %s", strings.Join(messages, "; "))
	}
	return nil
}

func jsonEqual(left, right json.RawMessage) bool {
	leftCanonical, leftErr := canonicalJSON(left)
	rightCanonical, rightErr := canonicalJSON(right)
	return leftErr == nil && rightErr == nil && leftCanonical == rightCanonical
}

func canonicalJSON(raw json.RawMessage) (string, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(value)
	return string(encoded), err
}

func cloneRaw(raw json.RawMessage) json.RawMessage {
	return append(json.RawMessage(nil), raw...)
}
