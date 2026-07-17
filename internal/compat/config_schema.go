package compat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// CheckConfigSchemaCompatibility enforces the supported-line compatibility
// policy for generated configuration schemas. Existing fields and accepted
// values must remain valid; new optional fields and relaxed enums are allowed.
func CheckConfigSchemaCompatibility(baseline, current []byte) error {
	oldSchema, err := decodeSchema("baseline", baseline)
	if err != nil {
		return err
	}
	newSchema, err := decodeSchema("current", current)
	if err != nil {
		return err
	}
	var breaks []string
	compareSchema("", oldSchema, newSchema, &breaks)
	if len(breaks) == 0 {
		return nil
	}
	sort.Strings(breaks)
	return fmt.Errorf("breaking config schema changes:\n- %s", strings.Join(breaks, "\n- "))
}

func decodeSchema(label string, raw []byte) (map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, fmt.Errorf("%s: invalid JSON Schema: %w", label, err)
	}
	root, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: JSON Schema root must be an object", label)
	}
	if props, exists := root["properties"]; exists {
		if _, ok := props.(map[string]any); !ok {
			return nil, fmt.Errorf("%s.properties: must be an object", label)
		}
	}
	return root, nil
}

func compareSchema(path string, oldSchema, newSchema map[string]any, breaks *[]string) {
	location := path
	if location == "" {
		location = "<root>"
	}

	for _, keyword := range []string{"type", "$ref", "const", "default", "allOf", "anyOf", "oneOf", "not"} {
		if oldValue, existed := oldSchema[keyword]; existed {
			if newValue, exists := newSchema[keyword]; !exists || !reflect.DeepEqual(oldValue, newValue) {
				*breaks = append(*breaks, fmt.Sprintf("%s changed %s", location, keyword))
			}
		} else if _, added := newSchema[keyword]; added && keyword != "type" {
			*breaks = append(*breaks, fmt.Sprintf("%s added restrictive %s", location, keyword))
		}
	}

	for _, keyword := range []string{"pattern", "format", "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf", "minLength", "maxLength", "minItems", "maxItems", "uniqueItems", "minProperties", "maxProperties", "additionalProperties"} {
		oldValue, oldOK := oldSchema[keyword]
		newValue, newOK := newSchema[keyword]
		if oldOK != newOK || (oldOK && !reflect.DeepEqual(oldValue, newValue)) {
			if newOK {
				*breaks = append(*breaks, fmt.Sprintf("%s changed constraint %s", location, keyword))
			}
		}
	}

	compareEnum(location, oldSchema["enum"], newSchema["enum"], breaks)
	compareRequired(location, oldSchema["required"], newSchema["required"], breaks)

	oldProperties, oldOK := oldSchema["properties"].(map[string]any)
	newProperties, newOK := newSchema["properties"].(map[string]any)
	if oldOK {
		if !newOK {
			*breaks = append(*breaks, fmt.Sprintf("%s removed properties", location))
			return
		}
		for name, oldValue := range oldProperties {
			childPath := name
			if path != "" {
				childPath = path + "." + name
			}
			newValue, exists := newProperties[name]
			if !exists {
				*breaks = append(*breaks, fmt.Sprintf("%s removed field", childPath))
				continue
			}
			oldChild, oldObject := oldValue.(map[string]any)
			newChild, newObject := newValue.(map[string]any)
			if !oldObject || !newObject {
				if !reflect.DeepEqual(oldValue, newValue) {
					*breaks = append(*breaks, fmt.Sprintf("%s changed schema", childPath))
				}
				continue
			}
			compareSchema(childPath, oldChild, newChild, breaks)
		}
	}

	oldItems, oldItemsOK := oldSchema["items"].(map[string]any)
	newItems, newItemsOK := newSchema["items"].(map[string]any)
	if oldItemsOK {
		if !newItemsOK {
			*breaks = append(*breaks, fmt.Sprintf("%s removed items schema", location))
		} else {
			compareSchema(path+"[]", oldItems, newItems, breaks)
		}
	}
}

func compareEnum(path string, oldValue, newValue any, breaks *[]string) {
	oldEnum, oldOK := oldValue.([]any)
	newEnum, newOK := newValue.([]any)
	if !oldOK {
		if newOK {
			*breaks = append(*breaks, fmt.Sprintf("%s added restrictive enum", path))
		}
		return
	}
	if !newOK {
		return
	}
	for _, oldItem := range oldEnum {
		found := false
		for _, newItem := range newEnum {
			if reflect.DeepEqual(oldItem, newItem) {
				found = true
				break
			}
		}
		if !found {
			*breaks = append(*breaks, fmt.Sprintf("%s removed enum value %v", path, oldItem))
		}
	}
}

func compareRequired(path string, oldValue, newValue any, breaks *[]string) {
	oldRequired := stringSet(oldValue)
	for field := range stringSet(newValue) {
		if _, existed := oldRequired[field]; !existed {
			childPath := field
			if path != "" && path != "<root>" {
				childPath = path + "." + field
			}
			*breaks = append(*breaks, fmt.Sprintf("%s became required", childPath))
		}
	}
}

func stringSet(value any) map[string]struct{} {
	result := map[string]struct{}{}
	items, _ := value.([]any)
	for _, item := range items {
		if text, ok := item.(string); ok {
			result[text] = struct{}{}
		}
	}
	return result
}
