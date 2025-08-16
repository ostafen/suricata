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

package spec

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Root of the spec
type Spec struct {
	Version  string             `yaml:"version"`
	Package  string             `yaml:"package"`
	Messages map[string]Message `yaml:"messages"`
	Tools    map[string]Tool    `yaml:"tools"`
	Agents   map[string]Agent   `yaml:"agents"`
}

type Message struct {
	Fields []Field `yaml:"fields"`
}

type Field struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description,omitempty"`
	Repeated    bool   `yaml:"repeated,omitempty"`
	Optional    bool   `yaml:"optional,omitempty"`
}

type Tool struct {
	Description string `yaml:"description"`
	Input       string `yaml:"input"`
	Output      string `yaml:"output"`
}

type Agent struct {
	Instructions string             `yaml:"instructions,omitempty"`
	Actions      map[string]Actions `yaml:"actions"`
	Tools        []string           `yaml:"tools"`
}

type Actions struct {
	Description string `yaml:"description"`
	Input       string `yaml:"input"`
	Output      string `yaml:"output"`
	Prompt      string `yaml:"prompt"`
	SkipInput   bool   `yaml:"skip_input"`
}

func LoadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}
	return &spec, spec.Validate()
}

// isPrimitiveType checks if the given type is a built-in primitive type
func isPrimitiveType(t string) bool {
	switch t {
	case "string", "int", "int32", "int64", "float", "float32", "float64", "bool", "datetime":
		return true
	default:
		return false
	}
}

func (spec *Spec) Validate() error {
	if spec.Version == "" {
		return fmt.Errorf("spec: version is required")
	}
	if spec.Package == "" {
		return fmt.Errorf("spec: package is required")
	}

	if err := spec.validateMessages(); err != nil {
		return err
	}

	if err := spec.validateTools(); err != nil {
		return err
	}
	return spec.validateAgents()
}

func (spec *Spec) validateMessages() error {
	for name, msg := range spec.Messages {
		if name == "" {
			return fmt.Errorf("spec: message has empty name")
		}
		for _, field := range msg.Fields {
			if field.Name == "" {
				return fmt.Errorf("spec: field in message %q has empty name", name)
			}
			if field.Type == "" {
				return fmt.Errorf("spec: field %q in message %q has empty type", field.Name, name)
			}
			// Validate field type existence
			if !isPrimitiveType(field.Type) {
				if _, ok := spec.Messages[field.Type]; !ok {
					return fmt.Errorf("spec: field %q in message %q references undefined type %q", field.Name, name, field.Type)
				}
			}
		}
	}
	return nil
}

func (spec *Spec) validateTools() error {
	for name, tool := range spec.Tools {
		if name == "" {
			return fmt.Errorf("spec: tool has empty name")
		}
		if tool.Input == "" {
			return fmt.Errorf("spec: tool %q missing input type", name)
		}
		if tool.Output == "" {
			return fmt.Errorf("spec: tool %q missing output type", name)
		}

		if _, ok := spec.Messages[tool.Input]; !ok {
			return fmt.Errorf("spec: tool %q input references undefined message %q", name, tool.Input)
		}
		if _, ok := spec.Messages[tool.Output]; !ok {
			return fmt.Errorf("spec: tool %q output references undefined message %q", name, tool.Output)
		}
	}
	return nil
}

func (spec *Spec) validateAgents() error {
	for name, agent := range spec.Agents {
		if name == "" {
			return fmt.Errorf("spec: agent has empty name")
		}

		for actionName, action := range agent.Actions {
			if actionName == "" {
				return fmt.Errorf("spec: agent %q has action with empty name", name)
			}
			if action.Input != "" {
				if _, ok := spec.Messages[action.Input]; !ok {
					return fmt.Errorf("spec: agent %q action %q input references undefined message %q", name, actionName, action.Input)
				}
			}
			if action.Output != "" {
				if _, ok := spec.Messages[action.Output]; !ok {
					return fmt.Errorf("spec: agent %q action %q output references undefined message %q", name, actionName, action.Output)
				}
			}
		}

		// Validate tools used by agent
		for _, toolName := range agent.Tools {
			if _, ok := spec.Tools[toolName]; !ok {
				return fmt.Errorf("spec: agent %q references undefined tool %q", name, toolName)
			}
		}
	}
	return nil
}
