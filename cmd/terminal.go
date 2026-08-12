package cmd

import (
	"fmt"
	"strings"
	"unicode"
)

// terminalSafeInline preserves printable text while rendering terminal control
// characters visibly. It is only for human-readable output; JSON and raw
// content modes must retain their existing byte semantics.
func terminalSafeInline(value string) string {
	var output strings.Builder
	output.Grow(len(value))

	for _, r := range value {
		switch r {
		case '\n':
			output.WriteString(`\n`)
		case '\r':
			output.WriteString(`\r`)
		case '\t':
			output.WriteString(`\t`)
		default:
			if unicode.IsPrint(r) {
				output.WriteRune(r)
				continue
			}

			switch {
			case r < 0x80:
				fmt.Fprintf(&output, `\x%02x`, r)
			case r <= 0xffff:
				fmt.Fprintf(&output, `\u%04x`, r)
			default:
				fmt.Fprintf(&output, `\U%08x`, r)
			}
		}
	}

	return output.String()
}

// terminalSafeBlock keeps intentional line breaks while ensuring continuation
// lines stay inside the caller-provided indentation. Bare carriage returns and
// other controls remain visible rather than active.
func terminalSafeBlock(value, continuationIndent string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	lines := strings.Split(value, "\n")
	for i := range lines {
		lines[i] = terminalSafeInline(lines[i])
	}
	return strings.Join(lines, "\n"+continuationIndent)
}
