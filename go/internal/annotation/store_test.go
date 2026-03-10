package annotation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAnnotationPath(t *testing.T) {
	got := AnnotationPath("/a/b/test.md")
	want := "/a/b/test.md.markcli.json"
	if got != want {
		t.Errorf("AnnotationPath() = %q, want %q", got, want)
	}
}

func TestLoad_fileExists(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")
	af := AnnotationFile{File: "test.md", Annotations: []Annotation{}}
	data, _ := json.Marshal(af)
	os.WriteFile(AnnotationPath(filePath), data, 0644)

	got, err := Load(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if got.File != "test.md" || len(got.Annotations) != 0 {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestLoad_fileNotExist(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")

	got, err := Load(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if got.File != "test.md" {
		t.Errorf("File = %q, want %q", got.File, "test.md")
	}
	if got.Annotations == nil || len(got.Annotations) != 0 {
		t.Errorf("expected empty annotations slice, got %v", got.Annotations)
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")
	af := AnnotationFile{File: "test.md", Annotations: []Annotation{}}

	if err := Save(filePath, af); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(AnnotationPath(filePath))
	if err != nil {
		t.Fatal(err)
	}
	// Verify it's valid JSON with correct content
	var got AnnotationFile
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.File != "test.md" {
		t.Errorf("File = %q, want %q", got.File, "test.md")
	}
}

func TestSave_indented(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")
	af := AnnotationFile{File: "test.md", Annotations: []Annotation{}}

	if err := Save(filePath, af); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(AnnotationPath(filePath))
	// Should be indented (contains newlines and spaces)
	content := string(data)
	if content == `{"file":"test.md","annotations":[]}` {
		t.Error("expected indented JSON, got compact")
	}
}

func TestClear_exists(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")
	os.WriteFile(AnnotationPath(filePath), []byte("{}"), 0644)

	if err := Clear(filePath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(AnnotationPath(filePath)); !os.IsNotExist(err) {
		t.Error("file should have been deleted")
	}
}

func TestClear_notExist(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")

	if err := Clear(filePath); err != nil {
		t.Errorf("Clear() on non-existent file should not error, got %v", err)
	}
}
