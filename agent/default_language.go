package agent

// Keep the historical operational rules separate from the output language.
// This is applied only to code-owned defaults, never to valid stored overrides.
func defaultEnglishOutput(body string) string {
	if body == "" {
		return body
	}
	return body + "\n\nOutput language: Write generated human-readable replies, summaries, and reports in English. This directive supersedes any Chinese-only output requirement above; all operational rules remain unchanged. Tool names, arguments, code, and quoted evidence must remain unchanged."
}
