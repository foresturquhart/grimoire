# Grimoire

Grimoire is a command-line tool that converts the contents of a directory into structured output formats optimized for interpretation by large language models (LLMs). It is lightweight, highly configurable, and user-friendly.

## Quick Start

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
```

## Features

* **Multiple Output Formats:** Generate output in Markdown, XML, or plain text formats to suit your needs.
* **Recursive File Scanning:** Automatically traverses directories and subdirectories to identify eligible files based on customizable extensions.
* **Content Filtering:** Skips ignored directories, temporary files, and patterns defined in the configuration.
* **Directory Tree Visualization:** Includes an optional directory structure representation at the beginning of the output.
* **Git Integration:** Prioritizes files by commit frequency when working within a Git repository.
* **Flexible Output:** Supports output to stdout or a specified file.

## Installation

### Prerequisites

* Git (required for repositories using Git-based sorting).

### Recommended: Download Pre-compiled Binary

The easiest way to install Grimoire is by downloading a pre-compiled binary from the [releases page](https://github.com/foresturquhart/grimoire/releases).

1. Visit the [releases page](https://github.com/foresturquhart/grimoire/releases).
2. Download the appropriate archive for your system (e.g., `grimoire-1.1.3-linux-amd64.tar.gz` or `grimoire-1.1.3-darwin-arm64.tar.gz`).
3. Extract the archive to retrieve the `grimoire` executable.
4. Move the `grimoire` executable to a directory in your system's `PATH` (e.g., `/usr/local/bin` or `~/.local/bin`). You may need to use `sudo` for system-wide locations:
   ```bash
   tar -xzf grimoire-1.1.3-linux-amd64.tar.gz
   cd grimoire-1.1.3-linux-amd64
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

## Configuration

### Allowed File Extensions

Grimoire processes files with specific extensions. You can customize these by modifying the `DefaultAllowedFileExtensions` constant in the codebase.

### Ignored Path Patterns

Files and directories matching patterns in the `DefaultIgnoredPathPatterns` constant are excluded from processing. This includes temporary files, build artifacts, and version control directories.

In addition, Grimoire honors ignore files located within the target directory. It supports both the traditional `.gitignore` and the new `.grimoireignore`. These files allow you to specify additional ignore rules on a per-directory basis. The rules defined in these files follow the same syntax as Git ignore rules, giving you fine-grained control over which files and directories should be omitted during the conversion process.

## Output Formats

Grimoire supports three output formats:

1. **Markdown (md)** - Default format that wraps file contents in code blocks with file paths as headings.
2. **XML (xml)** - Structures the content in an XML format with file paths as attributes.
3. **Plain Text (txt)** - Uses separator lines to distinguish between files.

Each format includes metadata, a summary section, an optional directory tree, and the content of all files.

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

## Feedback and Support

For issues, suggestions, or feedback, please open an issue on the [GitHub repository](https://github.com/foresturquhart/grimoire/issues).