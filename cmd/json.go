package cmd

import (
	"encoding/json"
	"fmt"
	"io"
)

func writeJSON(out io.Writer, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	_, err = fmt.Fprintln(out, string(data))
	return err
}
