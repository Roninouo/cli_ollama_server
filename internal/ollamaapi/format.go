package ollamaapi

import (
	"fmt"
	"strings"
	"time"
)

func FormatTags(models []TagModel) string {
	cols := []string{"NAME", "ID", "SIZE", "MODIFIED"}
	rows := make([][]string, 0, len(models))
	for _, m := range models {
		id := shortDigest(m.Digest)
		rows = append(rows, []string{m.Name, id, fmtBytes(m.Size), fmtTime(m.ModifiedAt)})
	}
	return formatTable(cols, rows)
}

func FormatPS(models []PSModel) string {
	cols := []string{"NAME", "ID", "SIZE", "UNTIL"}
	rows := make([][]string, 0, len(models))
	for _, m := range models {
		id := shortDigest(m.Digest)
		rows = append(rows, []string{m.Name, id, fmtBytes(m.Size), fmtTime(m.ExpiresAt)})
	}
	return formatTable(cols, rows)
}

func shortDigest(d string) string {
	d = strings.TrimSpace(d)
	if len(d) <= 12 {
		return d
	}
	return d[:12]
}

func fmtBytes(n int64) string {
	if n < 0 {
		return "?"
	}
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)
	switch {
	case n >= TB:
		return fmt.Sprintf("%.1f TB", float64(n)/float64(TB))
	case n >= GB:
		return fmt.Sprintf("%.1f GB", float64(n)/float64(GB))
	case n >= MB:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(MB))
	case n >= KB:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(KB))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func fmtTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	// Use a stable, locale-agnostic format.
	return t.UTC().Format("2006-01-02 15:04:05Z")
}

func formatTable(cols []string, rows [][]string) string {
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = len(c)
	}
	for _, r := range rows {
		for i := 0; i < len(cols) && i < len(r); i++ {
			if l := len(r[i]); l > widths[i] {
				widths[i] = l
			}
		}
	}

	var b strings.Builder
	for i, c := range cols {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(padRight(c, widths[i]))
	}
	b.WriteString("\n")

	for _, r := range rows {
		for i := 0; i < len(cols); i++ {
			if i > 0 {
				b.WriteString("  ")
			}
			v := ""
			if i < len(r) {
				v = r[i]
			}
			b.WriteString(padRight(v, widths[i]))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
