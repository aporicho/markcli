package annotation

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func AnnotationPath(filePath string) string {
	return filePath + ".markcli.json"
}

func Load(filePath string) (AnnotationFile, error) {
	annoPath := AnnotationPath(filePath)
	data, err := os.ReadFile(annoPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return AnnotationFile{File: filepath.Base(filePath), Annotations: []Annotation{}}, nil
		}
		return AnnotationFile{}, err
	}
	var af AnnotationFile
	if err := json.Unmarshal(data, &af); err != nil {
		return AnnotationFile{}, err
	}
	return af, nil
}

func Save(filePath string, af AnnotationFile) error {
	annoPath := AnnotationPath(filePath)
	data, err := json.MarshalIndent(af, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: temp file → chmod → rename
	dir := filepath.Dir(annoPath)
	tmp, err := os.CreateTemp(dir, ".markcli-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Chmod(tmpPath, 0600); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, annoPath)
}

func Clear(filePath string) error {
	annoPath := AnnotationPath(filePath)
	err := os.Remove(annoPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
