package embed

import (
	"fmt"
	"os"
)



func EmbedEnv(targetPath, callbackURL string) error {
	f, err := os.OpenFile(targetPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("cannnot open %s: %w", targetPath, err)
	}

	defer f.Close()
	_, err = fmt.Fprintf(f, "\nGOBAITR_CANARY=\"%s\"\n", callbackURL)
	return err
}