package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gopkg.in/yaml.v3"
)

// emailRequireSignedSender rides on GroupSettings, part of every group in
// AppData, which every mobile build deserializes at login. It must stay an
// optional boolean: a required field would fail payloads from older servers,
// and on UpdateGroupSettingsCommand an omitted key is what tells the server to
// leave the stored value alone.
func TestEmailRequireSignedSenderIsAnOptionalBooleanOnTheContract(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "..", "..", "swagger.yml"))
	if err != nil {
		t.Fatalf("read swagger.yml: %v", err)
	}

	var doc struct {
		Components struct {
			Schemas map[string]struct {
				Required   []string `yaml:"required"`
				Properties map[string]struct {
					Type string   `yaml:"type"`
					Enum []string `yaml:"enum"`
					Ref  string   `yaml:"$ref"`
				} `yaml:"properties"`
			} `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse swagger.yml: %v", err)
	}

	for _, name := range []string{"GroupSettings", "UpdateGroupSettingsCommand"} {
		t.Run(name, func(t *testing.T) {
			schema, ok := doc.Components.Schemas[name]
			if !ok {
				t.Fatalf("swagger.yml is missing the %s schema", name)
			}

			property, ok := schema.Properties["emailRequireSignedSender"]
			if !ok {
				t.Fatal("emailRequireSignedSender is not declared")
			}
			if property.Type != "boolean" || len(property.Enum) > 0 || property.Ref != "" {
				t.Errorf("emailRequireSignedSender must be a plain boolean, got type=%q enum=%v $ref=%q", property.Type, property.Enum, property.Ref)
			}
			for _, required := range schema.Required {
				if required == "emailRequireSignedSender" {
					t.Error("emailRequireSignedSender must stay optional")
				}
			}
		})
	}
}
