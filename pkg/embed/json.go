package embed

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func EmbedJSON(targetPath,callbackURL string) error {
	original, err := os.ReadFile(targetPath)
	if err != nil {
		return err
	}

	if !json.Valid(original) {
		return fmt.Errorf("target %s is not valid JSON -- cannot embed safely", targetPath) }

	backupPath := targetPath + ".gobaitr.bak"
	if err := os.WriteFile(backupPath, original, 0600); err != nil {
		return fmt.Errorf("failed to write backup %s: %w", backupPath, err)
	}

	indent := " "
	lines := strings.Split(string(original), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			indent = strings.TrimRight(line, strings.TrimLeft(line, " \t"))
			if indent == "" {
				indent = " "
			}
			break
		}
	}

	var data map[string]interface{}
	if err := json.Unmarshal(original, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	data["_gobaitr"] = callbackURL
	
	out, err := json.MarshalIndent(data, "", indent)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	writeErr := os.WriteFile(targetPath, out, 0600)
	if writeErr != nil {
		_ = os.WriteFile(targetPath, original, 0600)
		return fmt.Errorf("embed failed, original file restored: %w", writeErr)
	}
	return nil
}