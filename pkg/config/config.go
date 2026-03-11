package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Конфигурация линтера
type Config struct {
	Rules          RulesConfig          `yaml:"rules"`
	Settings       SettingsConfig       `yaml:"settings"`
	Loggers        LoggersConfig        `yaml:"loggers"`
	Global         GlobalConfig         `yaml:"global"`
	Autofix        AutofixConfig        `yaml:"autofix"`
	CustomPatterns CustomPatternsConfig `yaml:"custom_patterns"`
}

// Настройки кастомных паттернов
type CustomPatternsConfig struct {
	Enabled  bool                `yaml:"enabled"`
	Patterns []PatternConfig     `yaml:"patterns"`
}

// Конфигурация паттерна
type PatternConfig struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type"`
	Words       []string `yaml:"words"`
	Regex       string   `yaml:"regex"`
	Severity    string   `yaml:"severity"`
	Enabled     bool     `yaml:"enabled"`
	Tags        []string `yaml:"tags"`
}

// Настройки правил
type RulesConfig struct {
	LowercaseStart   bool `yaml:"lowercase_start"`
	EnglishOnly      bool `yaml:"english_only"`
	NoSpecialChars   bool `yaml:"no_special_chars"`
	NoSensitiveData  bool `yaml:"no_sensitive_data"`
}

// Детальные настройки правил
type SettingsConfig struct {
	LowercaseStart   LowercaseStartConfig   `yaml:"lowercase_start"`
	EnglishOnly      EnglishOnlyConfig      `yaml:"english_only"`
	NoSpecialChars   NoSpecialCharsConfig   `yaml:"no_special_chars"`
	NoSensitiveData  NoSensitiveDataConfig  `yaml:"no_sensitive_data"`
}

// Настройки правила строчной буквы
type LowercaseStartConfig struct {
	AllowedWords       []string `yaml:"allowed_words"`
	MaxLengthException int      `yaml:"max_length_exception"`
}

// Настройки проверки английского языка
type EnglishOnlyConfig struct {
	AllowedUnicode []string `yaml:"allowed_unicode"`
}

// Настройки спецсимволов
type NoSpecialCharsConfig struct {
	AllowedChars              []string `yaml:"allowed_chars"`
	BlockEmoji                bool     `yaml:"block_emoji"`
	BlockRepeatedPunctuation bool     `yaml:"block_repeated_punctuation"`
}

// Настройки чувствительных данных
type NoSensitiveDataConfig struct {
	SensitivePatterns []string `yaml:"sensitive_patterns"`
	SafeContexts      []string `yaml:"safe_contexts"`
}

// Настройки логгеров
type LoggersConfig struct {
	Supported          []string `yaml:"supported"`
	ExcludedFunctions  []string `yaml:"excluded_functions"`
}

// Глобальные настройки
type GlobalConfig struct {
	OutputFormat       string   `yaml:"output_format"`
	Verbosity          string   `yaml:"verbosity"`
	MaxErrorsPerFile   int      `yaml:"max_errors_per_file"`
	Exclude            ExcludeConfig `yaml:"exclude"`
}

// Настройки исключений
type ExcludeConfig struct {
	Patterns []string `yaml:"patterns"`
}

// Настройки авто-исправления
type AutofixConfig struct {
	Enabled bool     `yaml:"enabled"`
	Rules   []string `yaml:"rules"`
	Backup  bool     `yaml:"backup"`
}

// Конфигурация по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Rules: RulesConfig{
			LowercaseStart:  true,
			EnglishOnly:     true,
			NoSpecialChars:  true,
			NoSensitiveData: true,
		},
		Settings: SettingsConfig{
			LowercaseStart: LowercaseStartConfig{
				AllowedWords: []string{
					"API", "URL", "HTTP", "HTTPS", "JSON", "XML", "SQL", "DB", "ID", "UUID",
					"JWT", "OAuth", "REST", "GraphQL", "TCP", "UDP", "IP", "DNS", "TLS", "SSL",
				},
				MaxLengthException: 30,
			},
			EnglishOnly: EnglishOnlyConfig{
				AllowedUnicode: []string{"©", "®", "™"},
			},
			NoSpecialChars: NoSpecialCharsConfig{
				AllowedChars:              []string{".", ",", ":", "-"},
				BlockEmoji:                true,
				BlockRepeatedPunctuation: true,
			},
			NoSensitiveData: NoSensitiveDataConfig{
				SensitivePatterns: []string{
					"password", "passwd", "pwd", "api_key", "apikey", "api-key",
					"token", "secret", "key", "credential", "authorization",
					"bearer", "session", "cookie", "private_key", "private-key",
					"access_token", "refresh_token", "client_secret", "client-secret",
				},
				SafeContexts: []string{
					"token validated", "token received", "token processed", "token expired",
					"key updated", "key found", "key generated", "api request",
					"api response", "api call", "session started", "session ended",
					"session created", "session destroyed",
				},
			},
		},
		Loggers: LoggersConfig{
			Supported: []string{"slog", "log", "logger", "zap"},
			ExcludedFunctions: []string{"Print", "Printf", "Println"},
		},
		Global: GlobalConfig{
			OutputFormat:     "text",
			Verbosity:        "normal",
			MaxErrorsPerFile: 10,
			Exclude: ExcludeConfig{
				Patterns: []string{"*_test.go", "*.pb.go", "*.gen.go", "vendor/*", "mock/*", "testdata/*"},
			},
		},
		Autofix: AutofixConfig{
			Enabled: true,
			Rules:   []string{"lowercase_start", "english_only"},
			Backup:  true,
		},
		CustomPatterns: CustomPatternsConfig{
			Enabled:  true,
			Patterns: []PatternConfig{}, // Будут загружены из patterns пакет
		},
	}
}

// Загрузка конфигурации из файла
func LoadConfig(configPath string) (*Config, error) {
	// Если путь не указан, используем путь по умолчанию
	if configPath == "" {
		configPath = ".loglinter.yaml"
	}

	// Проверяем существование файла
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Если файл не существует, создаем конфигурацию по умолчанию
		fmt.Printf("Конфигурационный файл %s не найден, используется конфигурация по умолчанию\n", configPath)
		return DefaultConfig(), nil
	}

	// Читаем файл
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения конфигурационного файла %s: %w", configPath, err)
	}

	// Парсим YAML
	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфигурационного файла %s: %w", configPath, err)
	}

	// Валидация конфигурации
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("ошибка валидации конфигурации: %w", err)
	}

	return config, nil
}

// Валидация конфигурации
func validateConfig(config *Config) error {
	// Проверка формата вывода
	validFormats := []string{"text", "json", "yaml"}
	if !contains(validFormats, config.Global.OutputFormat) {
		return fmt.Errorf("недопустимый формат вывода: %s", config.Global.OutputFormat)
	}

	// Проверка уровня детализации
	validVerbosity := []string{"quiet", "normal", "verbose"}
	if !contains(validVerbosity, config.Global.Verbosity) {
		return fmt.Errorf("недопустимый уровень детализации: %s", config.Global.Verbosity)
	}

	// Проверка максимального количества ошибок
	if config.Global.MaxErrorsPerFile < 0 {
		return fmt.Errorf("максимальное количество ошибок не может быть отрицательным")
	}

	return nil
}

// Проверка наличия элемента в срезе
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Сохранение конфигурации в файл
func SaveConfig(config *Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("ошибка сериализации конфигурации: %w", err)
	}

	// Создаем директорию, если она не существует
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("ошибка создания директории %s: %w", dir, err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи конфигурационного файла %s: %w", configPath, err)
	}

	return nil
}
