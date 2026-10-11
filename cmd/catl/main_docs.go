//go:build docs

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

const docsOutputDir = "./docs/content/cli"

func main() {
	if err := generateDocs(newRootCommand()); err != nil {
		fmt.Fprintf(os.Stderr, "error generating docs: %v\n", err)
		os.Exit(1)
	}
}

// generateDocs cleans old .md files while strictly preserving _meta.json.
func generateDocs(root *cli.Command) error {
	if err := os.MkdirAll(docsOutputDir, 0o755); err != nil {
		return fmt.Errorf("create docs directory: %w", err)
	}

	entries, err := os.ReadDir(docsOutputDir)
	if err != nil {
		return fmt.Errorf("read docs directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			if err := os.Remove(filepath.Join(docsOutputDir, entry.Name())); err != nil {
				return fmt.Errorf("remove old %s: %w", entry.Name(), err)
			}
		}
	}

	return writeCommandDocs(root, nil)
}

// writeCommandDocs renders rich markdown for a command and its subcommands.
func writeCommandDocs(cmd *cli.Command, parentPath []string) error {
	if cmd.Hidden {
		return nil
	}

	currPath := append(append([]string(nil), parentPath...), cmd.Name)
	cmdName := strings.Join(currPath, " ")

	fileName := "catl.md"
	if len(currPath) > 1 {
		fileName = "catl-" + strings.Join(currPath[1:], "-") + ".md"
	}

	var buf bytes.Buffer

	// Frontmatter
	buf.WriteString(fmt.Sprintf("---\ntitle: %s\ndescription: %s\n---\n\n", cmdName, cmd.Usage))
	buf.WriteString("<!-- Generated automatically by catl docs. Do not edit directly. -->\n\n")

	// Single H1 title
	buf.WriteString(fmt.Sprintf("# %s\n\n", cmdName))
	buf.WriteString(fmt.Sprintf("%s\n\n", cmd.Usage))

	// Split description and examples block
	descBody, examples := splitDescriptionAndExamples(cmd.Description)
	if descBody != "" {
		buf.WriteString(fmt.Sprintf("%s\n\n", descBody))
	}

	// Metadata: Category & Aliases
	var metaLines []string
	if cmd.Category != "" {
		metaLines = append(metaLines, fmt.Sprintf("**Category:** %s", cmd.Category))
	}
	if len(cmd.Aliases) > 0 {
		var aliases []string
		for _, a := range cmd.Aliases {
			aliases = append(aliases, fmt.Sprintf("`catl %s`", a))
		}
		metaLines = append(metaLines, fmt.Sprintf("**Aliases:** %s", strings.Join(aliases, ", ")))
	}
	if len(metaLines) > 0 {
		buf.WriteString(strings.Join(metaLines, " \\\n") + "\n\n")
	}

	// Usage section (H2)
	buf.WriteString("## Usage\n\n```bash\n")
	if cmd.UsageText != "" {
		buf.WriteString(cmd.UsageText + "\n")
	} else if len(cmd.Commands) > 0 && len(currPath) == 1 {
		buf.WriteString("catl [flags] <command> [args...]\n")
	} else if len(cmd.Commands) > 0 {
		buf.WriteString(fmt.Sprintf("%s [flags] <subcommand> [args...]\n", cmdName))
	} else {
		buf.WriteString(fmt.Sprintf("%s [flags]\n", cmdName))
	}
	buf.WriteString("```\n\n")

	// Positional arguments table (H2)
	if len(cmd.Arguments) > 0 {
		buf.WriteString("## Arguments\n\n| Argument | Type | Required |\n| :--- | :--- | :--- |\n")
		for _, arg := range cmd.Arguments {
			usage, argType, req := formatArgDetails(arg)
			buf.WriteString(fmt.Sprintf("| `%s` | `%s` | %s |\n", usage, argType, req))
		}
		buf.WriteString("\n")
	}

	// Flags table (H2)
	if len(cmd.Flags) > 0 {
		buf.WriteString("## Flags\n\n| Flag | Description |\n| :--- | :--- |\n")
		for _, f := range cmd.Flags {
			synopsis, usage := formatFlag(f)
			buf.WriteString(fmt.Sprintf("| `%s` | %s |\n", synopsis, usage))
		}
		buf.WriteString("\n")
	}

	// Interactive Examples block (H2)
	if len(examples) > 0 {
		buf.WriteString("## Examples\n\n```bash\n")
		for _, ex := range examples {
			buf.WriteString(ex + "\n")
		}
		buf.WriteString("```\n\n")
	}

	// Subcommands table (H2)
	if len(cmd.Commands) > 0 {
		buf.WriteString("## Subcommands\n\n| Command | Description |\n| :--- | :--- |\n")
		for _, sub := range cmd.Commands {
			if sub.Hidden {
				continue
			}
			link := fmt.Sprintf("./catl-%s.md", sub.Name)
			if len(currPath) > 1 {
				link = fmt.Sprintf("./catl-%s-%s.md", strings.Join(currPath[1:], "-"), sub.Name)
			}
			buf.WriteString(fmt.Sprintf("| [`%s`](%s) | %s |\n", sub.Name, link, sub.Usage))
		}
		buf.WriteString("\n")
	}

	// Write file
	path := filepath.Join(docsOutputDir, fileName)
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	// Recurse subcommands
	for _, sub := range cmd.Commands {
		if err := writeCommandDocs(sub, currPath); err != nil {
			return err
		}
	}

	return nil
}

// formatArgDetails extracts name, type, and required status from positional arguments.
func formatArgDetails(arg cli.Argument) (usage, argType, required string) {
	switch ta := arg.(type) {
	case *cli.StringArg:
		req := "No"
		if ta.Required {
			req = "Yes"
		}
		name := ta.UsageText
		if name == "" {
			name = ta.Name
		}
		return name, "string", req
	case *cli.StringArgs:
		req := "No"
		if ta.Min > 0 {
			req = fmt.Sprintf("Yes (min %d)", ta.Min)
		}
		name := ta.UsageText
		if name == "" {
			name = ta.Name + "..."
		}
		return name, "string...", req
	default:
		return arg.Usage(), "string", "No"
	}
}

// splitDescriptionAndExamples splits Description into text and parsed examples.
func splitDescriptionAndExamples(desc string) (string, []string) {
	if desc == "" {
		return "", nil
	}

	idx := strings.Index(desc, "Examples:")
	if idx == -1 {
		return strings.TrimSpace(desc), nil
	}

	body := strings.TrimSpace(desc[:idx])
	rawExamples := desc[idx+len("Examples:"):]

	var examples []string
	for _, line := range strings.Split(rawExamples, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "$ ") {
			examples = append(examples, strings.TrimPrefix(trimmed, "$ "))
		} else if trimmed != "" {
			examples = append(examples, trimmed)
		}
	}

	return body, examples
}

// formatFlag formats flag names with short options first.
func formatFlag(f cli.Flag) (synopsis, usage string) {
	switch tf := f.(type) {
	case *cli.StringFlag:
		return flagNames(tf.Name, tf.Aliases, true), tf.Usage
	case *cli.BoolFlag:
		return flagNames(tf.Name, tf.Aliases, false), tf.Usage
	case *cli.StringSliceFlag:
		return flagNames(tf.Name, tf.Aliases, true), tf.Usage
	default:
		names := f.Names()
		if len(names) > 0 {
			return "--" + names[0], ""
		}
		return "", ""
	}
}

// flagNames formats flag names with short options first.
func flagNames(name string, aliases []string, hasArg bool) string {
	var parts []string
	for _, a := range aliases {
		if len(a) == 1 {
			parts = append(parts, "-"+a)
		} else {
			parts = append(parts, "--"+a)
		}
	}
	parts = append(parts, "--"+name)
	res := strings.Join(parts, ", ")
	if hasArg {
		res += " <value>"
	}
	return res
}
