package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func WriteJSON[T any](path string, data *T) error {
	if path == "" {
		return errors.New("path can't be empty")
	}
	if data == nil {
		return errors.New("data can't be nil")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create dir %s, %w", dir, err)
	}

	tmpFile, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file in %s, %w", dir, err)
	}
	tmpPath := tmpFile.Name()

	renamed := false
	defer func() {
		if !renamed {
			_ = os.Remove(tmpPath)
		}
	}()

	if err := tmpFile.Chmod(0o644); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to chmod temp file %s, %w", tmpPath, err)
	}

	enc := json.NewEncoder(tmpFile)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to encode data for %s, %w", path, err)
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to sync temp file %s, %w", tmpPath, err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file %s, %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename %s to %s, %w", tmpPath, path, err)
	}
	renamed = true

	return nil
}

func ReadJSON[T any](path string, data *T) error {
	if path == "" {
		return errors.New("path can't be empty")
	}
	if data == nil {
		return errors.New("data can't be nil")
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(data); err != nil {
		return fmt.Errorf("failed to decode %s, %w", path, err)
	}

	if dec.More() {
		return fmt.Errorf("unexpected trailing content after JSON value in %s", path)
	}

	return nil
}
