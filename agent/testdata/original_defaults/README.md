# Original built-in prompt fixtures

Source: `d0033724a63aefbbcc1ed022b0faf4dc4a1fce51`.

Extraction: parse the historical Go files with `go/parser`, locate each named
constant, recursively evaluate string literals and `+` expressions with
`strconv.Unquote`, and write the exact UTF-8 bytes without adding a newline.
This includes every concatenated literal in `ReporterDefaultPrompt`.

These fixtures pin the approved original operational rules, not model quality.
