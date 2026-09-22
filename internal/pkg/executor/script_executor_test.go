package executor

import "testing"

func TestValidateScriptType(t *testing.T) {
	tests := []struct {
		name         string
		scriptType   ScriptType
		allowedTypes []ScriptType
		want         bool
	}{
		{
			name:         "empty allowed list returns false",
			scriptType:   ScriptTypeBash,
			allowedTypes: []ScriptType{},
			want:         false,
		},
		{
			name:         "allowed type returns true",
			scriptType:   ScriptTypeJavaScript,
			allowedTypes: []ScriptType{ScriptTypeJavaScript, ScriptTypePython},
			want:         true,
		},
		{
			name:         "disallowed type returns false",
			scriptType:   ScriptTypeBash,
			allowedTypes: []ScriptType{ScriptTypeJavaScript},
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateScriptType(tt.scriptType, tt.allowedTypes); got != tt.want {
				t.Errorf("ValidateScriptType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterEnvVars(t *testing.T) {
	env := []string{
		"PATH=/usr/bin",
		"HOME=/tmp",
		"JWT_SECRET=secret",
		"DB_PASSWORD=password",
		"LC_ALL=C",
	}

	got := filterEnvVars(env)
	for _, item := range got {
		if item == "JWT_SECRET=secret" || item == "DB_PASSWORD=password" {
			t.Fatalf("filterEnvVars leaked sensitive variable: %s", item)
		}
	}

	if len(got) != 3 {
		t.Fatalf("filterEnvVars() length = %d, want 3; got %#v", len(got), got)
	}
}
