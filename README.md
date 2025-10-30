# go-hml

go-hml is a small, fast command-line utility for counting source code lines and comment lines by file extension. It recursively scans a directory tree, filters files by extension and ignore list, parses files in parallel using one worker per logical CPU, and prints a tidy, human-friendly table (or a compact single-line summary in quiet mode). The tool is written in Go and designed to be simple, fast, and easy to configure.

## Features

* Recursively walk a project directory and collect files by extension.
* Filter files by an ignore list (directory or filename).
* Include only specified extensions (optional).
* Count code lines and single-line comments (customizable comment prefixes).
* Parallel file parsing (workers = logical CPUs).
* Pretty table output or compact quiet output.
* Colorized extension column by default (can be disabled with `NO_COLOR`).

## Installation

Build the binary from the repository root:

```bash
go build -o hml .
```

You can also `go install` it into your `GOBIN`:

```bash
go install ./...
```

## Usage

Basic invocation:

```bash
./hml <path>
```

Examples:

```bash
# Scan a project, include Go/JS/Python files and ignore vendor and .git
./hml -e "go,js,py" -i "vendor,.git" ./project

# Use a JSON config file
./hml --config=config.json ./project

# Quiet mode (only totals)
./hml -q ./project
```

### Command-line flags

* `-i` / `--ignore` — comma-separated list of directories or filenames to ignore (for example `vendor,.git,node_modules`).
* `-e` / `--extensions` — comma-separated list of file extensions to include (without dot), e.g. `go,js,py`. If omitted, all files with extensions are considered.
* `-c` / `--comments` — comma-separated list of single-line comment prefixes to detect (default: `//`). Examples: `//,#`.
* `-q` / `--quiet` — output only the final numbers (compact single-line format).
* `--config` — path to a JSON config file (see below).
* `--help` — show usage help.

If a setting is present both in the JSON config and as a command-line flag, the command-line flag takes precedence.

## JSON configuration

You can provide a JSON file with the same options:

```json
{
  "ignore_list": "vendor,.git,node_modules",
  "extensions": "go,js,py",
  "comments_list": "//,#",
  "quiet": false
}
```

Pass it with `--config=config.json`. Values from the config are applied only when the corresponding flag was not passed on the command line.

## Output

Normal table output lists extension, files, code lines and comment lines and ends with totals, for example:

```
+-----------+-------+------+----------+
| Extension | Files | Code | Comments |
+-----------+-------+------+----------+
| go        |   12  | 3456 |   789    |
| js        |    8  | 1024 |   200    |
+-----------+-------+------+----------+
Total: Files=20  Code=4480  Comments=989
```

Quiet mode prints a single compact line:

```
547 (code: 4480 comments: 989)
```

Files without an extension are skipped. If the environment variable `NO_COLOR` is set, colorized output is disabled.

## Implementation notes and where to contribute

The core logic lives in the internal package that walks directories, filters files, parses each file line-by-line and aggregates per-extension results. File parsing currently treats a line as a comment when it begins with any of the configured single-line comment prefixes; multi-line comment support and more advanced comment detection are upgradable areas. The worker pool size is `max(1, runtime.NumCPU())`, which gives good throughput on multi-core machines.

If you want to extend or harden the tool, good first contributions are tests for `ParseFile`, adding support for multi-line comments and block comment detection, improving language-specific comment heuristics, and adding a subcommand or flag to output JSON/CSV machine-readable reports.

## Logging and error handling

Access errors to individual files or directories are logged to stderr and scanning continues where possible. Fatal file read/parsing errors will terminate with a non-zero exit code. The tool aims to be resilient: non-fatal errors are printed but do not stop the entire run.
