// Copyright (c) 2025 Suricata Contributors
// Original Author: Stefano Scafiti
//
// This file is part of Suricata: Type-Safe AI Agents for Go.
//
// Licensed under the MIT License. You may obtain a copy of the License at
//
//	https://opensource.org/licenses/MIT
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gen

import (
	"fmt"

	"github.com/ostafen/suricata/pkg/spec"
)

type JSONSchema map[string]any

type JSONSchemaGenerator struct {
	schemas map[string]JSONSchema
}

func NewJSONSchemaGenerator() *JSONSchemaGenerator {
	return &JSONSchemaGenerator{
		schemas: make(map[string]JSONSchema),
	}
}

// GenerateJSONSchema returns a JSON Schema object (as a map) for the given message.
// It recursively includes referenced custom types.
func (gen *JSONSchemaGenerator) GenerateJSONSchema(name string, msg *spec.Message, allMessages map[string]spec.Message) (JSONSchema, error) {
	schema, has := gen.schemas[name]
	if has {
		return schema, nil
	}

	schema, err := gen.generateJSONSchema(msg, allMessages)
	if err != nil {
		return nil, err
	}

	gen.schemas[name] = schema
	return schema, nil
}

func (gen *JSONSchemaGenerator) generateJSONSchema(msg *spec.Message, allMessages map[string]spec.Message) (JSONSchema, error) {
	properties := make(map[string]any)

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	requiredFields := []string{}
	for _, field := range msg.Fields {
		fieldSchema, err := gen.fieldToSchema(field, allMessages)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", field.Name, err)
		}

		properties[field.Name] = fieldSchema

		// If not optional, add to required
		if !field.Optional {
			requiredFields = append(requiredFields, field.Name)
		}
	}

	if len(requiredFields) > 0 {
		schema["required"] = requiredFields
	}
	return schema, nil
}

// fieldToSchema generates the JSON Schema for a single field, recursively if needed.
func (gen *JSONSchemaGenerator) fieldToSchema(field spec.Field, allMessages map[string]spec.Message) (map[string]interface{}, error) {
	var baseSchema map[string]any

	switch field.Type {
	case "string":
		baseSchema = map[string]any{"type": "string"}
	case "int", "int32", "int64":
		baseSchema = map[string]any{"type": "integer"}
	case "float", "float32", "float64":
		baseSchema = map[string]any{"type": "number"}
	case "bool":
		baseSchema = map[string]any{"type": "boolean"}
	case "datetime":
		baseSchema = map[string]any{"type": "string", "format": "date-time"} // RFC3339
	default:
		// Custom type - lookup in allMessages
		msg, ok := allMessages[field.Type]
		if !ok {
			return nil, fmt.Errorf("unknown custom type %q", field.Type)
		}

		// Recursive schema for nested message
		nestedSchema, err := gen.GenerateJSONSchema(field.Type, &msg, allMessages)
		if err != nil {
			return nil, err
		}
		baseSchema = nestedSchema
	}

	if field.Description != "" {
		baseSchema["description"] = field.Description
	}

	// Wrap in array if repeated
	if field.Repeated {
		return map[string]any{
			"type":  "array",
			"items": baseSchema,
		}, nil
	}
	return baseSchema, nil
}
