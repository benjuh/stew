<img width="642" alt="image" src="https://github.com/benjuh/stew/assets/82689821/94145b53-e0e2-4beb-b9ad-a34fae888875">

Introducing `stew` 🍲🎉. A CLI for creating, storing, and using templates to reduce boilerplate.

## Installation

Install with Go:

```
go install github.com/benjuh/stew@latest
```

Prebuilt binaries for macOS, Linux, and Windows are available on the
[GitHub releases page](https://github.com/benjuh/stew/releases). Download the
archive for your operating system and architecture, then place `stew` on your
`PATH`.

`stew upgrade` updates Go-installed copies with `go install` and verifies and
replaces release-binary installations on macOS and Linux. Homebrew and other
package-manager installations should be upgraded through their package manager.

Upgrade an existing installation from inside the CLI:

```bash
stew upgrade
stew upgrade --check
stew upgrade v1.6.0
```

Generate shell completion scripts:

```bash
stew completion zsh > ~/.zsh/completions/_stew
stew completion bash > ~/.local/share/bash-completion/completions/stew
stew completion fish > ~/.config/fish/completions/stew.fish
```

## Usage
- `stew save`: save an existing directory as a template (`add` remains an alias)
- `stew edit`: edit the values of a saved stew 
- `stew view`: view saved template metadata and its project tree (`get` remains an alias)
- `stew list`: list all saved stew templates
- `stew search`: search a remote starter catalog
- `stew browse`: browse a remote starter catalog
- `stew info`: inspect a catalog template and its variants
- `stew import`: turn an archive or Git project into a local template
- `stew create`: create a project from a template (`new` remains an alias)
- `stew remove`: remove a stew template
- `stew replace`: replace all instances of a string in a project
- `stew doctor`: diagnose a project, manifest, and task configuration
- `stew tasks`: list tasks available in the current project
- `stew run`: run a named project task
- `stew upgrade`: upgrade Stew to the latest or a specific version
- `stew completion`: generate Bash, Zsh, Fish, or PowerShell completions

By default, `stew create` will not overwrite existing files. Use `--force` only
when overwriting files in the destination is intended. Missing destination
directories are created automatically.

Templates can use Go template expressions in text files and paths. Supply
values with repeatable `--var` flags:

```bash
stew create react-app my-dashboard \
  --var project_name=my-dashboard \
  --var module=github.com/me/my-dashboard
```

For example, `README.md` can contain `# {{ .ProjectName }}` and a directory
can be named `{{ .ProjectName }}`. Variable names are also available in their
original form, such as `{{ .project_name }}`. Use `--dry-run` to inspect the
planned files or `--json` for machine-readable output. Existing files are
protected unless `--force` is supplied.

Templates may include a `.stewignore` file with simple file, directory, or
glob patterns. `.git`, `node_modules`, `vendor`, and other VCS directories are
ignored automatically.

Templates may also include an optional `.stew.yaml` manifest containing a
description, tags, and variable documentation. Search templates with
`stew list --search react` or filter them with `stew list --tag frontend`.

Templates can extend another saved template with `extends: base-template`.
Child files override parent files, while variables, tasks, and tags are
merged. Generated projects receive a standalone merged manifest.

Manifests can also define project tasks. Generated projects retain the
manifest, so tasks can be run from the project directory:

```yaml
tasks:
  fmt:
    description: Format the project
    command: go
    args: [fmt, ./...]
    detect: [go.mod]
    aliases: [format]
  check:
    depends_on: [fmt, test]
    command: go
    args: [test, ./...]
  test:
    command: go
    args: [test, ./...]
    timeout: 5m
```

Use `stew tasks` to list available tasks and `stew run fmt` to run one.
`stew run fmt --dry-run` previews the command, while `--yes` skips the
confirmation prompt. Go, Rust, and Node projects also receive useful built-in
tasks when their standard project files are present. Tasks may use `aliases`,
`depends_on`, per-platform command overrides under `platforms`, environment
variables, a project-relative `dir`, and duration-based `timeout` values.

Run `stew doctor` before sharing or creating from a template. It checks the
manifest, renderable files, task dependencies, working directories, timeouts,
and task executables. Use `stew doctor --json` for CI diagnostics.

Templates can be shared as portable archives or installed from Git:

```bash
stew export react-app ./react-app.tar.gz
stew import ./react-app.tar.gz
stew install https://github.com/example/templates.git react-app
stew update react-app
stew outdated
stew diff react-app
stew verify react-app
```

Curated starter templates can be discovered from a catalog index:

```bash
stew search react --catalog-url https://benjuh.com/stew/catalog/index.json
stew search react --refresh
stew info react-neon-render --catalog-url https://benjuh.com/stew/catalog/index.json
stew install react-neon-render --variant minimal \
  --catalog-url https://benjuh.com/stew/catalog/index.json --refresh
stew create react-neon-render --variant minimal
```

Catalog entries describe neutral `minimal` and more complete `standard`
variants. The catalog contains metadata and sources; template files remain in
the separate template repository.

Remote catalog results are cached briefly for offline use. Add `--refresh` to
`search`, `browse`, `info`, or variant-based `install` when you need the latest
catalog immediately. Empty catalogs are not cached.

Installed templates are cached under `~/.config/stew/templates` by default.
Archives exclude `.git` metadata, preserve file permissions, and reject unsafe
paths during extraction.

Git-backed templates record their resolved commit. Updates refuse to run when
the cached template has local modifications, and `--check` or `--dry-run` can
inspect updates without changing template files.

Manifest variables can define defaults and required values:

```yaml
variables:
  - name: project_name
    description: Name of the project
    required: true
  - name: author
    default: Benjamin
```

When creating a project, values can come from `--var`, a YAML values file, or
environment variables. Missing required values are prompted for interactively:

```bash
stew create react-app my-app --values project.yaml
stew create react-app my-app --non-interactive
stew validate react-app
```

For more information on usage, checkout the [DOC.md](https://github.com/benjuh/stew/blob/main/DOC.md) file.

## Configuration

The current options for configuration are:

- `stewsPath`: the path to the file where stews are stored (default: `$HOME/.stews.json`)
- `timeFormat`: the format for the time that stews are created (default: `2006-01-02 15:04:05`)
- `templatesPath`: cache location for imported and installed templates (default: `$HOME/.config/stew/templates`)
- `catalogURL`: remote starter catalog index (default: `https://benjuh.com/stew/catalog/index.json`)

Here is an example of a basic configuration file you could make up:

```yaml
stewsPath: /home/ben/.stews.json
timeFormat: Jan 2, 2006 @ 3:04pm
catalogURL: https://benjuh.com/stew/catalog/index.json
```

If you have any suggestions or issues, feel free to open an issue or PR. Enjoy! 🎉

## Examples of Use Cases
Personally, there are certain templates I like to start with for markdown files, react projects, etc., so why not create a way to easily create, store and use templates from within your terminal?

A good way to do this is to have some folder of templates you like for example in `$HOME/.config/stew/templates` or `$HOME/Documents/templates` and add new stews directed toward those paths.

For example, say you make a templates directory at `$HOME/Documents/templates` and you made a basic markdown template for when you take notes or configure a README at `$Home/Documents/templates/markdown`. You can then use `stew save markdown -p $HOME/Documents/templates/markdown -d "basic README layout"` to save it in your templates and later create a project with `stew create markdown`.

You may be wondering how this is any different from running a simple `cp -r dir1 dir2` and it comes down to organization and ease of use. I simply prefer being able to run `stew list` and see all my created stews and be able to set descriptions and configure it the way I want than to just use `cp -r`. I plan to add many more features to take this above and beyond but I need help coming up with those ideas. So if you have any requests please create an Issue or contribute yourself by cloning the repo and submitting a PR. I will do my best to review promptly.
