package env

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/akpatel363/menv/internal/config"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

func TestParseEnvFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string]string
	}{
		{
			name:    "simple key value",
			content: "FOO=bar\nBAZ=qux\n",
			want:    map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name:    "double quoted value strips quotes",
			content: `FOO="hello world"`,
			want:    map[string]string{"FOO": "hello world"},
		},
		{
			name:    "single quoted value strips quotes",
			content: `FOO='hello world'`,
			want:    map[string]string{"FOO": "hello world"},
		},
		{
			name:    "export prefix is stripped",
			content: "export FOO=bar\n",
			want:    map[string]string{"FOO": "bar"},
		},
		{
			name:    "comments and blank lines are skipped",
			content: "# header\n\nFOO=bar\n# trailing\n",
			want:    map[string]string{"FOO": "bar"},
		},
		{
			name:    "value containing equals is preserved",
			content: "URL=https://example.com/?a=1&b=2\n",
			want:    map[string]string{"URL": "https://example.com/?a=1&b=2"},
		},
		{
			name:    "later definition wins",
			content: "FOO=one\nFOO=two\n",
			want:    map[string]string{"FOO": "two"},
		},
	}

	dir := t.TempDir()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := writeFile(t, dir, "env."+tc.name, tc.content)
			got, err := parseEnvFile(path)
			if err != nil {
				t.Fatalf("parseEnvFile: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLoadEnv_OverridePrecedence(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".env.base", "FOO=base\nBAR=base\n")
	writeFile(t, dir, ".env.local", "BAR=local\nBAZ=local\n")

	project := config.Project{Path: dir}
	envCfg := config.Env{
		Files:     []string{".env.base", ".env.local"},
		Overrides: map[string]string{"BAZ": "override"},
	}

	got, err := LoadEnv(project, envCfg)
	if err != nil {
		t.Fatalf("LoadEnv: %v", err)
	}

	want := map[string]string{
		"FOO": "base",     // only in base
		"BAR": "local",    // local overrides base
		"BAZ": "override", // overrides beats both files
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestLoadEnv_MissingFileErrors(t *testing.T) {
	project := config.Project{Path: t.TempDir()}
	envCfg := config.Env{Files: []string{"nope.env"}}

	if _, err := LoadEnv(project, envCfg); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestBuildEnv_LoadedOverridesOSEnv(t *testing.T) {
	t.Setenv("MENV_TEST_KEY", "from-os")
	merged := BuildEnv(map[string]string{"MENV_TEST_KEY": "from-loaded"})

	var found string
	for _, kv := range merged {
		if len(kv) > len("MENV_TEST_KEY=") && kv[:len("MENV_TEST_KEY=")] == "MENV_TEST_KEY=" {
			found = kv[len("MENV_TEST_KEY="):]
		}
	}
	if found != "from-loaded" {
		t.Errorf("MENV_TEST_KEY = %q, want %q", found, "from-loaded")
	}
}
