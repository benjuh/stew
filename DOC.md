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
```

Examples:

```bash
stew search react --catalog-url https://benjuh.com/stew/catalog/index.json
stew browse --catalog-url ./catalog.yaml
stew info react-neon-render --json --catalog-url ./catalog.yaml
```

Catalog entries describe a template family, language, tags, and variants. The
CLI does not embed starter files; entries point to sources in a separate
template repository.

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
Import a template archive into the local catalog

Usage:
  stew import <archive> [flags]

Flags:
      --name string  Name to register the imported template under
```

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

`remove` deletes only the catalog entry; it does not delete the template
directory or cached files.

### Replace a project name

```bash
stew replace old-service new-service --path ./my-project
stew replace Acme acme --ignore-case
```

Replacement skips binary files and common dependency/VCS directories.

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
