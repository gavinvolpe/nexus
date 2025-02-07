
## Available Prompts for Selection {{range .}}

### `{{.Id}}` - {{.Title}}
  Target: {{.Target}}
  When to use: {{.When}}

  #### Prompt:
    “{{.Prompt}}”

  #### Purpose:
    {{.Purpose}}
{{end}}