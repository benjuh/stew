# Commands

## save (add alias)
```
Save an existing directory as a template

Usage:
stew save <name_of_stew> [flags]

Flags:
-d, --description string   Description of the template (default "no description provided")
  -h, --help                 help for save
-p, --path string          Path to the template (defaults to current directory)
```

## edit
```
Edit the values of a saved stew

Usage:
stew edit <name_of_stew> [flags]

Required Flags:
at least one of --n , -p, or -d must be provided to make an edit

Flags:
  -d, --description string   The new description of the stew
  -h, --help                 help for edit
  -n, --name string          The new name of the stew
  -p, --path string          The new path of the stew

Global Flags:
      --config string   config file (default is $HOME/.config/stew/config.yaml)
```

## view (get alias)
```
View values and the project tree of a saved stew

Usage:
  stew view <name_of_stew> [flags]

Flags:
  -h, --help          help for view
  -t, --tree          Print the tree (retained for compatibility)
      --no-tree       Do not print the template tree
      --depth int     Limit tree depth; zero means unlimited
      --json          Output template details and tree as JSON

Global Flags:
      --config string   config file (default is $HOME/.config/stew/config.yaml)
```

## list
```
List saved templates

Usage:
  stew list [flags]

Flags:
  -h, --help   help for list
      --search string  Search template names, descriptions, and tags
      --tag string     Only show templates with this tag
      --json           Output templates as JSON

Global Flags:
  --config string   config file (default is $HOME/.config/stew/config.yaml)
```

## search, browse, and info
```
Discover templates from a remote or local catalog index.

Usage:
  stew search [query] [flags]
  stew browse [flags]
  stew info <template> [flags]

Flags:
      --catalog-url string   Catalog URL or local index file
      --json                 Output JSON
      --refresh              Fetch the catalog instead of using the fresh cache
```

Examples:

```bash
stew search react --catalog-url https://benjuh.com/stew/catalog/index.json
stew search react --refresh
stew browse --catalog-url ./catalog.yaml
stew info react-neon-render --json --catalog-url ./catalog.yaml
stew install react-neon-render --variant minimal --refresh
```

Catalog entries describe a template family, language, tags, and variants. The
CLI does not embed starter files; entries point to sources in a separate
template repository.

Remote catalogs are cached briefly for offline use. Use `--refresh` with
`search`, `browse`, `info`, or variant-based `install` to fetch the latest
catalog immediately. Empty catalogs are never cached.

### create (new alias)
```
Create a project from a template

Usage:
stew create <name_of_stew> [destination] [flags]

Flags:
  -h, --help          help for create
  -p, --path string       Destination directory (defaults to current directory)
  -f, --force             Overwrite existing files
      --var stringArray   Template variable in key=value form (repeatable)
      --variant string    Catalog variant previously installed for this template
      --dry-run           Show planned changes without writing files
      --json              Output planned or completed changes as JSON
      --values string     YAML file containing template variables
      --non-interactive   Fail instead of prompting for missing required variables

  Global Flags:
--config string   config file (default is $HOME/.config/stew/config.yaml)
```

## export
```
Export a template as a portable archive

Usage:
  stew export <name_of_stew> <archive> [flags]
```

## import
```
Import an archive or curate a Git project into the local catalog

Usage:
  stew import <archive-or-git-url> [flags]

Flags:
      --description string  Description for an imported Git template
      --dry-run             Show what would be imported without saving it
      --exclude strings     Additional file or directory names to exclude
      --name string         Name to register the imported template under
      --profile string      Curation profile: auto, generic, react, node, or go
```

Import a Git project as a curated Stew template:

```bash
stew import https://github.com/example/react-project \
  --name react-starter --profile react
stew import https://github.com/example/service --profile go --dry-run
```

Git imports clone into a temporary directory, remove generated files using a
profile, create a minimal `.stew.yaml` when one is missing, and register the
curated result locally. Existing manifests are preserved. Use `--exclude` for
additional file or directory names.

## install
```
Install a template from a directory, archive, or Git URL

Usage:
  stew install <source> [name] [flags]

Flags:
      --ref string          Git branch, tag, or commit to install
      --variant string      Catalog variant to install
      --catalog-url string  Catalog URL or local index file
```

Install a catalog variant by template ID:

```bash
stew install react-neon-render --variant minimal \
  --catalog-url https://benjuh.com/stew/catalog/index.json
```

When no name is supplied, the installed name is `<template-id>-<variant>`.
Create that installed variant with `stew create <template-id> --variant <variant>`.

## update
```
Update a Git-backed template

Usage:
  stew update <name_of_stew> [flags]

Flags:
      --ref string    Git branch, tag, or commit to update to
      --check         Check for updates without changing files
      --dry-run       Show the update without changing files
```

## outdated
```
Check installed Git templates for updates

Usage:
  stew outdated [flags]

Flags:
      --json          Output update status as JSON
```

## upgrade
```
Upgrade Stew to the latest or a specific version

Usage:
  stew upgrade [version] [flags]

Flags:
      --check   Show the latest available version without upgrading
```

Examples:

```bash
stew upgrade
stew upgrade --check
stew upgrade v1.6.0
```

`upgrade` uses `go install` for Go-installed copies. Release binaries on macOS
and Linux are downloaded from GitHub Releases and verified against the
published checksums before replacement. Package-manager installations should
be upgraded through their package manager instead. Windows users should
download the new archive manually because a running Windows executable cannot
replace itself safely.

Prebuilt archives for macOS, Linux, and Windows are published on the GitHub
releases page. Package-manager installations should be upgraded through their
package manager instead of replacing the managed binary manually.

## completion
```
Generate shell completion scripts

Usage:
  stew completion <shell>

Supported shells:
  bash, zsh, fish, powershell
```

Examples:

```bash
mkdir -p ~/.zsh/completions
stew completion zsh > ~/.zsh/completions/_stew

mkdir -p ~/.config/fish/completions
stew completion fish > ~/.config/fish/completions/stew.fish
```

Restart the shell or reload its completion configuration after installing a
generated script.

## diff
```
Show changes available for a Git-backed template

Usage:
  stew diff <name_of_stew> [flags]
```

## verify
```
Verify a template source, checksum, and manifest

Usage:
  stew verify <name_of_stew> [flags]
```

## remove
```
Remove a stew

Usage:
  stew remove <name_of_stew> [flags]

Flags:
  -h, --help          help for remove
      --keep-files    Keep template files on disk; only remove the catalog entry

Global Flags:
  --config string   config file (default is $HOME/.config/stew/config.yaml)
```

## validate
```
Validate a template and its manifest

Usage:
  stew validate <name_of_stew> [flags]

Flags:
  -h, --help          help for validate
  -p, --path string   Validate this directory instead of the saved template path
```

## tasks
```
List tasks available in the current project.

Usage:
  stew tasks [flags]

Flags:
      --json   Output tasks as JSON
```

## run
```
Run a named project task from the nearest project root.

Usage:
  stew run <task> [flags]

Flags:
      --dry-run   Show the task without executing it
      --yes       Skip confirmation
```

Task manifests support `description`, `aliases`, `depends_on`, `detect`, and
`platforms`. Dependencies run once in dependency order before the requested
task. Cycles and missing tasks are rejected before execution.

## doctor
```
Diagnose the current project and its tasks.

Usage:
  stew doctor [flags]

Flags:
      --json          Output diagnostics as JSON
      --path string   Project directory to diagnose (defaults to current directory)
```

Task runtime settings include `env`, a project-relative `dir`, and a duration
`timeout` such as `30s` or `5m`.

Catalog variants may include a relative `path` so multiple starters can live
in one Git repository. The CLI clones the repository to a temporary staging
directory, copies only that path into the managed template cache, and removes
the staging directory.

## replace
```
Replace all instances of a string in a project

Usage:
stew replace <old_string> <new_string> [flags]

Flags:
  -h, --help          help for replace
  -p, --path string   The path to the stew (defaults to current directory)
  -i, --ignore-case   Ignore case when searching for the old string

Replacement skips binary files and common dependency/VCS directories, replaces
literal text, and preserves file permissions.

  Global Flags:
--config string   config file (default is $HOME/.config/stew/config.yaml)
```

## Examples and use cases

### Start a local template

```bash
stew save go-service --path ./templates/go-service \
  --description "Go service starter"
stew view go-service --depth 2
stew validate go-service
stew create go-service ./billing-service \
  --var project_name=billing-service \
  --var module=example.com/billing-service
```

Use `save` when the template is already on your machine. Use `install` when
you want stew to copy and manage a template from a directory, archive, or Git
source.

### Search and inspect templates

```bash
stew list --search backend
stew list --tag go --json
stew view go-service --json
```

Use `list` for catalog discovery and `view` for one template’s metadata and
tree. `get` remains an alias for `view`.

### Create safely in automation

```bash
stew create go-service ./build \
  --values ci-values.yaml \
  --non-interactive \
  --dry-run \
  --json
```

Remove `--dry-run` when the plan is approved. Existing files are protected by
default; add `--force` only when overwriting is expected.

### Diagnose and run a generated project

```bash
cd ./build
stew doctor
stew doctor --json > doctor-report.json
stew tasks
stew run fmt --dry-run
stew run test --yes
```

`validate` checks a saved template’s rendering. `doctor` checks a project’s
manifest, renderability, task dependencies, executables, working directories,
and timeouts.

### Share a template as an archive

```bash
stew export go-service ./go-service.tar.gz
stew import ./go-service.tar.gz --name go-service-copy
stew create go-service-copy ./new-service
```

Use `export`/`import` for a portable file. Archives preserve permissions,
exclude VCS metadata, and reject unsafe paths.

### Install and maintain a Git template

```bash
stew install https://github.com/example/templates.git api --ref main
stew outdated
stew diff api
stew update api --check
stew update api
stew verify api
```

Use `outdated` to find updates, `diff` to inspect them, `update --check` or
`--dry-run` to preview them, and `verify` to check recorded source integrity.
Updates refuse to overwrite locally modified cached templates.

### Manage catalog entries

```bash
stew edit go-service --description "Updated Go service starter"
stew edit go-service --path ~/templates/go-service
stew remove go-service
```

`remove` deletes the catalog entry and removes templates stored under Stew's
managed cache. It preserves template directories stored elsewhere. Use
`--keep-files` to remove only the catalog entry.

### Replace a project name

```bash
stew replace old-service new-service --path ./my-project
stew replace Acme acme --ignore-case
```

Replacement skips binary files and common dependency/VCS directories.

## Complete template workflow

This walkthrough creates a reusable Go service template with variables,
ignored files, task dependencies, platform-specific behavior, and a generated
project workflow. The same pattern works for React, Node, Rust, and other
project types by changing the files and commands.

### 1. Create the template files

Create a directory and add a manifest:

```bash
mkdir -p ~/templates/go-service/{cmd/service,internal}
cd ~/templates/go-service
```

Create `.stew.yaml`:

```yaml
description: Go HTTP service starter
tags: [go, backend, service]

variables:
  - name: project_name
    description: Display name used in documentation
    required: true
  - name: module
    description: Go module path
    required: true
  - name: author
    description: Owner shown in generated documentation
    default: Example Team

tasks:
  setup:
    description: Download Go dependencies
    command: go
    args: [mod, download]
    detect: [go.mod]
    timeout: 5m
  fmt:
    description: Format Go source files
    command: go
    args: [fmt, ./...]
    aliases: [format]
    detect: [go.mod]
  test:
    description: Run the test suite
    command: go
    args: [test, ./...]
    detect: [go.mod]
    env:
      CGO_ENABLED: "0"
    timeout: 5m
  check:
    description: Format and test the service
    depends_on: [fmt, test]
    command: go
    args: [vet, ./...]
    detect: [go.mod]

  # On Windows, use go.exe for the same task.
  # Other platforms use the command and args above.
  platform-check:
    description: Run the platform-specific checker
    command: go
    args: [vet, ./...]
    detect: [go.mod]
    platforms:
      windows:
        command: go.exe
        args: [vet, ./...]
```

Add a Go module whose path is rendered from the variables:

```go
// go.mod
module {{ .module }}

go 1.22
```

Add a source file and documentation:

```go
// cmd/service/main.go
package main

import "net/http"

func main() {
	_ = http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{{ .project_name }} is running\n"))
	}))
}
```

````markdown
# {{ .project_name }}

Owned by {{ .author }}.

Module: `{{ .module }}`

## Development

```bash
stew run setup --yes
stew run fmt --yes
stew run check --yes
```
````

Keep machine-specific and generated files out of the template with
`.stewignore`:

```text
.env
.env.*
bin/
coverage/
tmp/
```

Stew also ignores common dependency and VCS directories such as `.git`,
`node_modules`, and `vendor` automatically.

### 2. Save and validate the template

Register the directory in the local catalog and inspect what Stew sees:

```bash
stew save go-service --path ~/templates/go-service \
  --description "Go HTTP service starter"
stew list --tag go
stew view go-service --depth 3
stew validate go-service
```

`view` shows the saved metadata and project tree. `validate` checks the
manifest and attempts to render template files using the declared variables.
Fix validation errors before sharing or creating from the template.

### 3. Preview and create a project

Use a values file when several values belong together:

```yaml
# billing-service.yaml
project_name: Billing Service
module: example.com/acme/billing
author: Acme Platform Team
```

Preview the planned files without writing anything:

```bash
stew create go-service ./billing-service \
  --values ./billing-service.yaml \
  --dry-run
```

Create the project after reviewing the preview:

```bash
stew create go-service ./billing-service \
  --values ./billing-service.yaml \
  --non-interactive
cd billing-service
```

Values are resolved in this order: environment variables, manifest defaults,
the `--values` YAML file, `--var` flags, and finally interactive prompts.
Therefore, a one-off override can be supplied without editing the values file:

```bash
stew create go-service ./billing-service \
  --values ./billing-service.yaml \
  --var author="Release Engineering"
```

Use `--json` when another script needs to consume the planned or completed
file changes. Existing files are protected unless `--force` is supplied.

### 4. Diagnose and run the generated workflow

Generated projects retain the merged `.stew.yaml`, so the workflow travels
with the project:

```bash
stew doctor
stew tasks
stew run setup --yes
stew run fmt --yes
stew run check --dry-run
stew run check --yes
```

`check` runs its dependencies once in dependency order before running `go
vet`. `detect: [go.mod]` makes Go-specific tasks unavailable in projects that
do not contain a Go module. `doctor --json` is useful for CI diagnostics:

```bash
stew doctor --json > doctor-report.json
```

### 5. Share, import, and maintain the template

For a portable file, export the saved template and import it elsewhere:

```bash
stew export go-service ./go-service.tar.gz
stew import ./go-service.tar.gz --name go-service-copy
stew create go-service-copy ./another-service \
  --var project_name="Another Service" \
  --var module=example.com/acme/another
```

For a Git-backed template, install it directly and keep it updated:

```bash
stew install https://github.com/example/templates.git go-service --ref main
stew outdated
stew diff go-service
stew update go-service --check
stew update go-service
stew verify go-service
```

When starting from a full existing project, import it with a curation profile
instead of manually copying dependency folders and secrets:

```bash
stew import https://github.com/example/full-go-service \
  --name curated-go-service --profile go --dry-run
stew import https://github.com/example/full-go-service \
  --name curated-go-service --profile go \
  --exclude docs/generated
```

The imported template is stored in Stew's managed cache. `stew remove` removes
both its catalog entry and managed files; use `stew remove --keep-files` when
you only want to unregister it.

## Manifest reference

Templates may contain `.stew.yaml` at their root. It is copied into generated
projects.

### Inheritance

A template can extend one other template registered in the local catalog:

```yaml
extends: go-service
description: Go service with PostgreSQL
tags: [go, postgres]
```

Parent templates are resolved first, then the child is layered on top. Child
files override parent files at the same path. Variables and tasks with the
same name are replaced by the child; new ones are added. Tags are combined
without duplicates, and a non-empty child description replaces the parent
description.

Inheritance can be chained, but cycles and missing parents are errors. The
generated project receives the fully merged manifest with `extends` removed,
so it remains independent of the creator’s catalog.

```yaml
description: Go service starter
tags: [go, backend]
variables:
  - name: project_name
    required: true
  - name: author
    default: Example Team
tasks:
  fmt:
    description: Format the project
    command: go
    args: [fmt, ./...]
    aliases: [format]
    detect: [go.mod]
  check:
    depends_on: [fmt, test]
    command: go
    args: [vet, ./...]
  test:
    command: go
    args: [test, ./...]
    env:
      CGO_ENABLED: "0"
    dir: backend
    timeout: 5m
```

Task fields are `command`, `args`, `description`, `aliases`, `depends_on`,
`detect`, `platforms`, `env`, project-relative `dir`, and duration-based
`timeout` values such as `30s` or `5m`. Tasks execute without shell
interpolation and dependencies run once in dependency order.

Starter templates commonly expose a `setup` task for dependency installation.
Run it explicitly after creating a project:

```bash
stew run setup --yes
```

Variable values come from environment variables, manifest defaults, a
`--values` YAML file, `--var` flags, and finally interactive prompts, in that
precedence order. Use `.stewignore` for files that should not be copied.

## Configuration and global flags

Every command accepts the global configuration flag:

```text
      --config string   Config file (default: $HOME/.config/stew/config.yaml)
  -v, --version         Print the version
```

Default configuration values are:

```yaml
stewsPath: ~/.stews.json
timeFormat: "2006-01-02 15:04:05"
templatesPath: ~/.config/stew/templates
catalogURL: https://benjuh.com/stew/catalog/index.json
```

- `stewsPath`: catalog file containing saved templates.
- `timeFormat`: timestamp display format.
- `templatesPath`: managed cache for imported and installed templates.
- `catalogURL`: optional default remote or local catalog index.
