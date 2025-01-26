# Grimoire

Grimoire is a command-line tool that converts the contents of a directory into a structured Markdown document, making it suitable for interpretation by large language models (LLMs). It is designed to be lightweight, highly configurable, and easy to use.

## Features

- **Recursive File Scanning**: Scans directories and subdirectories for eligible files based on customizable extensions.
- **Content Filtering**: Skips ignored directories, temporary files, and patterns specified in the configuration.
- **Markdown Conversion**: Outputs structured Markdown with file headings and content wrapped in triple-backtick code blocks.
- **Git Integration**: Sorts files by commit frequency for relevance-based ordering if scanning a Git repository.
- **Flexible Output**: Write output to stdout or a specified file.

## Installation

### Prerequisites

- Git (required for repositories using Git sorting)

### Download Binary (Recommended)

Download the pre-compiled binary for your platform from the [releases page](https://github.com/foresturquhart/grimoire/releases).

### Install using Go

You can install Grimoire using Go:

```bash
go install github.com/foresturquhart/grimoire/cmd/grimoire@latest
```

### Build from Source

Alternatively, you can build from source:

1. Clone the repository:
   ```bash
   git clone https://github.com/foresturquhart/grimoire.git
   cd grimoire
   ```

2. Build the binary:
   ```bash
   go build -o grimoire ./cmd/grimoire
   ```

3. Move the binary to your PATH:
   ```bash
   mv grimoire /usr/local/bin/
   ```

4. Verify installation:
   ```bash
   grimoire --version
   ```

## Usage

### Basic Command

```bash
grimoire [options] <target directory>
```

### Options

- `-o, --output <path>`: Specify the output file. Defaults to stdout if not provided.
- `-f, --force`: Overwrite the output file if it already exists.
- `--version`: Display the current version of the tool.

### Examples

1. Convert a directory into Markdown and print to stdout:
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

4. Use in a Git repository and sort files by commit frequency:
   ```bash
   grimoire ./my-git-repo
   ```

## Configuration

### Allowed Extensions

Grimoire processes files with specific extensions by default. You can customize the allowed extensions by modifying the `AllowedExtensions` constant in the codebase.

### Ignored Patterns

Files and directories matching the `IgnoredPatterns` constant are skipped during processing. This includes temporary files, build artifacts, and version control directories.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bugfix:
   ```bash
   git checkout -b feature/my-new-feature
   ```
3. Commit your changes:
   ```bash
   git commit -m "Add my new feature"
   ```
4. Push to your branch:
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

If you encounter issues or have suggestions, please open an issue on the [GitHub repository](https://github.com/foresturquhart/grimoire/issues).