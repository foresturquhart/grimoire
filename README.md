<div align="center">
  <a href="https://github.com/foresturquhart/grimoire">
    <img src=".github/grimoire.png" alt="Grimoire" width="500" height="auto" />
  </a>
</div>

<hr />

[![go](https://badgen.net/static/go/1.24.3)](https://go.dev/)
[![license](https://badgen.net/github/license/foresturquhart/grimoire)](https://github.com/foresturquhart/grimoire/blob/main/LICENSE)
[![release](https://badgen.net/github/release/foresturquhart/grimoire/stable)](https://github.com/foresturquhart/grimoire/releases)

Grimoire is a command-line tool that converts the contents of a directory into structured output formats optimized for interpretation by Large Language Models (LLMs) like Claude, ChatGPT, Gemini, DeepSeek, Llama, Grok, and more. It is lightweight, highly configurable, and user-friendly.

## Quick Start

### Install Grimoire

The fastest way to get started is with our one-line installation script:

```bash
curl -sSL https://raw.githubusercontent.com/foresturquhart/grimoire/main/install.sh | bash
```

This script automatically detects your OS and architecture, downloads the appropriate binary, and installs it to your PATH.

### Basic Usage

```bash
# Convert current directory to Markdown and copy to clipboard (macOS)
grimoire . | pbcopy

# Convert current directory to Markdown and copy to clipboard (Linux with xclip)
grimoire . | xclip -selection clipboard

# Convert current directory and save to file
grimoire -o output.md .

# Convert current directory to XML format
grimoire --format xml -o output.xml .

# Convert current directory to plain text format
grimoire --format txt -o output.txt .

# Convert directory with secret detection and redaction
grimoire --redact-secrets -o output.md .
```

## Features

* **Multiple Output Formats:** Generate output in Markdown, XML, or plain text formats to suit your needs.
* **Recursive File Scanning:** Automatically traverses directories and subdirectories to identify eligible files based on customizable extensions.
* **Content Filtering:** Skips ignored directories, temporary files, and patterns defined in the configuration.
* **Directory Tree Visualization:** Includes an optional directory structure representation at the beginning of the output.
* **Git Integration:** Prioritizes files by commit frequency when working within a Git repository.
* **Secret Detection:** Scans files for potential secrets or sensitive information to prevent accidental exposure.
* **Secret Redaction:** Optionally redacts detected secrets in the output while preserving the overall code structure.
* **Token Counting:** Calculates the token count of generated output to help manage LLM context limits.
* **Minified File Detection:** Automatically identifies minified JavaScript and CSS files to warn about high token usage.
* **Flexible Output:** Supports output to stdout or a specified file.

## Installation

### Prerequisites

* Git (required for repositories using Git-based sorting).

### Quickest: One-line Installation Script

The easiest way to install Grimoire is with our automatic installation script:

```bash
curl -sSL https://raw.githubusercontent.com/foresturquhart/grimoire/main/install.sh | bash
```

This script automatically:
- Detects your operating system and architecture
- Downloads the appropriate binary for your system
- Installs it to `/usr/local/bin` (or `~/.local/bin` if you don't have sudo access)
- Makes the binary executable

### Alternative: Download Pre-compiled Binary

You can also manually download a pre-compiled binary from the [releases page](https://github.com/foresturquhart/grimoire/releases).

1. Visit the [releases page](https://github.com/foresturquhart/grimoire/releases).
2. Download the appropriate archive for your system (e.g., `grimoire-1.2.1-linux-amd64.tar.gz` or `grimoire-1.2.1-darwin-arm64.tar.gz`).
3. Extract the archive to retrieve the `grimoire` executable.
4. Move the `grimoire` executable to a directory in your system's `PATH` (e.g., `/usr/local/bin` or `~/.local/bin`). You may need to use `sudo` for system-wide locations:
   ```bash
   tar -xzf grimoire-1.2.1-linux-amd64.tar.gz
   cd grimoire-1.2.1-linux-amd64
   sudo mv grimoire /usr/local/bin/
   ```
5. Verify the installation:
   ```bash
   grimoire --version
   ```

### Install using `go install`

For users with Go installed, `go install` offers a straightforward installation method:

```bash
go install github.com/foresturquhart/grimoire/cmd/grimoire@latest
```

### Build from Source

To build Grimoire from source (useful for development or customization):

1. Clone the repository:
   ```bash
   git clone https://github.com/foresturquhart/grimoire.git
   cd grimoire
   ```
2. Build the binary:
   ```bash
   go build -o grimoire ./cmd/grimoire
   ```
3. Move the binary to your `PATH`:
   ```bash
   mv grimoire /usr/local/bin/
   ```
4. Verify the installation:
   ```bash
   grimoire --version
   ```

## Usage

### Basic Command

```bash
grimoire [options] <target directory>
```

### Options

- `-o, --output <path>`: Specify an output file. Defaults to stdout if omitted.
- `-f, --force`: Overwrite the output file if it already exists.
- `--format <format>`: Specify the output format. Options are `md` (or `markdown`), `xml`, and `txt` (or `text`, `plain`, `plaintext`). Defaults to `md`.
- `--no-tree`: Disable the directory tree visualization at the beginning of the output.
- `--no-sort`: Disable sorting files by Git commit frequency.
- `--ignore-secrets`: Proceed with output generation even if secrets are detected.
- `--redact-secrets`: Redact detected secrets in output rather than failing.
- `--skip-token-count`: Skip counting output tokens.
- `--version`: Display the current version.

### Examples

1. Convert a directory into Markdown and print the output to stdout:
   ```bash
   grimoire ./myproject
   ```
2. Save the output to a file:
   ```bash
   grimoire -o output.md ./myproject
   ```
3. Overwrite an existing file:
   ```bash
   grimoire -o output.md -f ./myproject
   ```
4. Generate XML output without a directory tree:
   ```bash
   grimoire --format xml --no-tree -o output.xml ./myproject
   ```
5. Generate plain text output without Git-based sorting:
   ```bash
   grimoire --format txt --no-sort -o output.txt ./myproject
   ```
6. Scan for secrets and redact them in the output:
   ```bash
   grimoire --redact-secrets -o output.md ./myproject
   ```

## Configuration

### Allowed File Extensions

Grimoire processes files with specific extensions. You can customize these by modifying the `DefaultAllowedFileExtensions` constant in the codebase.

### Ignored Path Patterns

Files and directories matching patterns in the `DefaultIgnoredPathPatterns` constant are excluded from processing. This includes temporary files, build artifacts, and version control directories.

### Custom Ignore Files

Grimoire supports two types of ignore files to specify additional exclusion patterns:

1. **`.gitignore`**: Standard Git ignore files are honored if present in the target directory.
2. **`.grimoireignore`**: Grimoire-specific ignore files that follow the same syntax as Git ignore files.

These files allow you to specify additional ignore rules on a per-directory basis, giving you fine-grained control over which files and directories should be omitted during the conversion process.

### Large File Handling

By default, Grimoire warns when processing files larger than 1MB. These files are still included in the output, but a warning is logged to alert you about potential performance impacts when feeding the output to an LLM.

### Minified File Detection

Grimoire automatically detects minified JavaScript and CSS files using several heuristics:
- Excessive line length
- Low ratio of lines to characters
- Presence of coding patterns typical in minified files

When a minified file is detected, Grimoire logs a warning, as these files can consume a large number of tokens while providing limited value to the LLM.

## Output Formats

Grimoire supports three output formats:

1. **Markdown (md)** - Default format that wraps file contents in code blocks with file paths as headings.
2. **XML (xml)** - Structures the content in an XML format with file paths as attributes.
3. **Plain Text (txt)** - Uses separator lines to distinguish between files.

Each format includes metadata, a summary section, an optional directory tree, and the content of all files.

## Token Counting

Grimoire includes built-in token counting to help you manage LLM context limits. The token count is estimated using the same tokenizer used by many LLMs. You can disable token counting entirely using the `--skip-token-count` flag.

## Secret Detection

Grimoire includes built-in secret detection powered by [gitleaks](https://github.com/gitleaks/gitleaks) to help prevent accidentally sharing sensitive information when using the generated output with LLMs or other tools.

By default, Grimoire scans for a wide variety of potential secrets including:

- API keys and tokens (AWS, GitHub, GitLab, etc.)
- Private keys (RSA, SSH, PGP, etc.)
- Authentication credentials
- Service-specific tokens (Stripe, Slack, Twilio, etc.)

The secret detection behavior can be controlled with the following flags:

- `--ignore-secrets`: Continues with output generation even if secrets are detected (logs warnings)
- `--redact-secrets`: Automatically redacts any detected secrets with the format `[REDACTED SECRET: description]`

If a secret is detected and neither of the above flags are specified, Grimoire will abort the operation and display a warning message, helping prevent accidental exposure of sensitive information.

## Contributing

Contributions are welcome! To get started:

1. Fork the repository.
2. Create a new branch for your feature or fix:
   ```bash
   git checkout -b feature/my-new-feature
   ```
3. Commit your changes:
   ```bash
   git commit -m "Add my new feature"
   ```
4. Push the branch to your fork:
   ```bash
   git push origin feature/my-new-feature
   ```
5. Open a pull request.

## License

Grimoire is licensed under the [MIT License](LICENSE).

## Acknowledgements

Grimoire uses the following libraries:

- [zerolog](https://github.com/rs/zerolog) for structured logging.
- [gitleaks](https://github.com/zricethezav/gitleaks) for secret detection and scanning.
- [go-gitignore](https://github.com/sabhiram/go-gitignore) for handling ignore patterns.
- [urfave/cli](https://github.com/urfave/cli) for command-line interface.
- [tiktoken-go](https://github.com/tiktoken-go/tokenizer) for token counting.

## Feedback and Support

For issues, suggestions, or feedback, please open an issue on the [GitHub repository](https://github.com/foresturquhart/grimoire/issues).