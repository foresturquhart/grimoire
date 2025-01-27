package config

// DefaultAllowedFileExtensions defines the default file extensions that are eligible for processing.
// These extensions represent common programming, configuration, and documentation file types.
var DefaultAllowedFileExtensions = []string{
	// Programming languages
	"rs", "c", "h", "cpp", "hpp", "py", "java", "go", "rb", "php", "cs",
	"fs", "fsx", "fsi", "fsscript", "scala", "kt", "kts", "dart", "swift",
	"m", "mm", "r", "pl", "pm", "t", "lua", "elm", "erl", "ex", "exs", "zig",
	"psgi", "cgi", "groovy",

	// Web and frontend files
	"html", "css", "sass", "scss", "js", "ts", "jsx", "tsx", "vue", "svelte",
	"haml", "hbs", "jade", "less", "coffee", "astro",

	// Configuration and data files
	"toml", "json", "yaml", "yml", "ini", "conf", "cfg", "properties", "env",
	"xml", "sql", "htaccess",

	// Documentation and markup
	"md", "mdx", "markdown", "txt", "graphql", "proto", "prisma", "dhall",

	// Build and project files
	"gitignore", "lock", "gradle", "pom", "sbt", "gemspec", "podspec", "rake",

	// Infrastructure files
	"sh", "fish", "tf", "tfvars",
}

// DefaultIgnoredPathPatterns defines the default path patterns that are excluded from processing.
// These include directories, build artifacts, caches, and temporary files.
var DefaultIgnoredPathPatterns = []string{
	// Common directories to ignore
	`^\.git/`, `^\.next/`, `^node_modules/`, `^vendor/`, `^dist/`,
	`^build/`, `^out/`, `^target/`, `^bin/`, `^obj/`, `^coverage/`,
	`^test-results/`, `^\.idea/`, `^\.vscode/`, `^\.vs/`, `^\.settings/`,
	`^\.gradle/`, `^\.mvn/`, `^\.pytest_cache/`, `^__pycache__/`,
	`^\.sass-cache/`, `^\.vercel/`, `^\.turbo/`,

	// Lock files and dependency metadata
	`pnpm-lock\.yaml`, `package-lock\.json`, `yarn\.lock`, `Cargo\.lock`,
	`Gemfile\.lock`, `composer\.lock`, `mix\.lock`, `poetry\.lock`,
	`Pipfile\.lock`, `packages\.lock\.json`, `paket\.lock`,

	// Temporary and binary files
	`\.pyc$`, `\.pyo$`, `\.pyd$`, `\.class$`, `\.o$`, `\.obj$`,
	`\.dll$`, `\.exe$`, `\.so$`, `\.dylib$`, `\.log$`, `\.tmp$`,
	`\.temp$`, `\.swp$`, `\.swo$`, `\.bak$`, `~$`,

	// System files
	`\.DS_Store$`, `Thumbs\.db$`, `\.env(\..+)?$`,

	// Specific files
	`^LICENSE$`, `\.gitignore`,
}
