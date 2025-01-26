# Grimoire

Grimoire is a command-line tool that converts the contents of a directory into a structured Markdown document, making it suitable for interpretation by large language models (LLMs). It is lightweight, highly configurable, and user-friendly.

## Features

* **Recursive File Scanning:** Automatically traverses directories and subdirectories to identify eligible files based on customizable extensions.
* **Content Filtering:** Skips ignored directories, temporary files, and patterns defined in the configuration.
* **Markdown Conversion:** Outputs structured Markdown with file headings and content enclosed in triple-backtick code blocks.
* **Git Integration:** Prioritizes files by commit frequency when working within a Git repository.
* **Flexible Output:** Supports output to stdout or a specified file.

## Installation

### Prerequisites

* Git (required for repositories using Git-based sorting).

### Recommended: Download Pre-compiled Binary

The easiest way to install Grimoire is by downloading a pre-compiled binary from the [releases page](https://github.com/foresturquhart/grimoire/releases).

1. Visit the [releases page](https://github.com/foresturquhart/grimoire/releases).
2. Download the appropriate archive for your system (e.g., `grimoire-1.1.0-linux-amd64.tar.gz` or `grimoire-1.1.0-darwin-arm64.tar.gz`).
3. Extract the archive to retrieve the `grimoire` executable.
4. Move the `grimoire` executable to a directory in your system's `PATH` (e.g., `/usr/local/bin` or `~/.local/bin`). You may need to use `sudo` for system-wide locations:
   ```bash
   tar -xzf grimoire-1.1.0-linux-amd64.tar.gz
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
4. Sort files by commit frequency in a Git repository:
   ```bash
   grimoire ./my-git-repo
   ```

## Configuration

### Allowed File Extensions

Grimoire processes files with specific extensions. You can customize these by modifying the `DefaultAllowedFileExtensions` constant in the codebase.

### Ignored Path Patterns

Files and directories matching patterns in the `DefaultIgnoredPathPatterns` constant are excluded from processing. This includes temporary files, build artifacts, and version control directories.

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

- [gocodewalker](https://github.com/boyter/gocodewalker) for efficient file traversal.
- [zerolog](https://github.com/rs/zerolog) for structured logging.

## Feedback and Support

For issues, suggestions, or feedback, please open an issue on the [GitHub repository](https://github.com/foresturquhart/grimoire/issues).