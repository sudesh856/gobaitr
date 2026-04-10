package embed

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestEmbedEnv(t *testing.T) {
	f, _ := os.CreateTemp("", "test-*.env")
	f.WriteString("EXISTING=value\n")
	f.Close()
	defer os.Remove(f.Name())

	err := EmbedEnv(f.Name(), "http://localhost/t/abc/secret")
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(f.Name())
	content := string(data)
	if !strings.Contains(content, "GOBAITR_CANARY") {
		t.Error("missing GOBAITR_CANARY line")
	}
	if !strings.Contains(content, "EXISTING=value") {
		t.Error("original lines must be preserved")
	}
}

func TestEmbedJSON(t *testing.T) {
	f, _ := os.CreateTemp("", "test-*.json")
	f.WriteString(`{"key": "value"}`)
	f.Close()
	defer os.Remove(f.Name())

	err := EmbedJSON(f.Name(), "http://localhost/t/abc/secret")
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(f.Name())
	if !strings.Contains(string(data), "_gobaitr") {
		t.Error("missing _gobaitr key")
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		t.Error("result is not valid JSON")
	}
}

func TestEmbedJSONInvalid(t *testing.T) {
	f, _ := os.CreateTemp("", "test-invalid-*.json")
	original := `{invalid json`
	f.WriteString(original)
	f.Close()
	defer os.Remove(f.Name())

	err := EmbedJSON(f.Name(), "http://localhost/t/abc/secret")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}

	data, _ := os.ReadFile(f.Name())
	if string(data) != original {
		t.Error("file was modified despite invalid JSON — Patch #8 violated")
	}
}

func TestEmbedText(t *testing.T) {
	f, _ := os.CreateTemp("", "test-*.txt")
	f.WriteString("existing content\n")
	f.Close()
	defer os.Remove(f.Name())

	err := EmbedText(f.Name(), "http://localhost/t/abc/secret")
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(f.Name())
	if !strings.Contains(string(data), "gobaitr") {
		t.Error("missing gobaitr comment")
	}
}

func TestEmbedJSONBackup(t *testing.T) {
	f, _ := os.CreateTemp("", "test-backup-*.json")
	original := `{"key": "value"}`
	f.WriteString(original)
	f.Close()
	defer os.Remove(f.Name())
	defer os.Remove(f.Name() + ".bak")

	err := EmbedJSON(f.Name(), "http://localhost/t/abc/secret")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(f.Name() + ".gobaitr.bak"); os.IsNotExist(err) {
		t.Error("backup file was not created before write")
	}

	bak, _ := os.ReadFile(f.Name() + ".gobaitr.bak")
	if string(bak) != original {
		t.Error("backup does not match original content")
	}
}
