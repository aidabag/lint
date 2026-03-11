package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Тест загрузки конфигурации по умолчанию
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	// Проверяем значения по умолчанию
	if !config.Rules.LowercaseStart {
		t.Error("LowercaseStart должен быть включен по умолчанию")
	}
	if !config.Rules.EnglishOnly {
		t.Error("EnglishOnly должен быть включен по умолчанию")
	}
	if !config.Rules.NoSpecialChars {
		t.Error("NoSpecialChars должен быть включен по умолчанию")
	}
	if !config.Rules.NoSensitiveData {
		t.Error("NoSensitiveData должен быть включен по умолчанию")
	}
	
	// Проверяем настройки по умолчанию
	if len(config.Settings.LowercaseStart.AllowedWords) == 0 {
		t.Error("Должны быть разрешенные слова по умолчанию")
	}
	if config.Settings.LowercaseStart.MaxLengthException != 30 {
		t.Error("MaxLengthException должен быть 30 по умолчанию")
	}
}

// Тест загрузки конфигурации из файла
func TestLoadConfig(t *testing.T) {
	// Создаем временный конфигурационный файл
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")
	
	configContent := `
rules:
  lowercase_start: false
  english_only: true
  no_special_chars: true
  no_sensitive_data: false

settings:
  lowercase_start:
    allowed_words:
      - "CUSTOM"
      - "TEST"
    max_length_exception: 50
  
  no_sensitive_data:
    sensitive_patterns:
      - "custom_secret"
      - "private_data"
    safe_contexts:
      - "custom context"
      - "test context"

global:
  output_format: "json"
  verbosity: "verbose"
  max_errors_per_file: 20
  exclude:
    patterns:
      - "*.custom.go"
      - "test_exclude/*"

autofix:
  enabled: false
  rules:
    - "english_only"
  backup: false
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Ошибка создания конфигурационного файла: %v", err)
	}
	
	// Загружаем конфигурацию
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	
	// Проверяем загруженные значения
	if config.Rules.LowercaseStart {
		t.Error("LowercaseStart должен быть отключен")
	}
	if !config.Rules.EnglishOnly {
		t.Error("EnglishOnly должен быть включен")
	}
	if config.Rules.NoSensitiveData {
		t.Error("NoSensitiveData должен быть отключен")
	}
	
	// Проверяем настройки
	if len(config.Settings.LowercaseStart.AllowedWords) != 2 {
		t.Error("Должно быть 2 разрешенных слова")
	}
	if config.Settings.LowercaseStart.MaxLengthException != 50 {
		t.Error("MaxLengthException должен быть 50")
	}
	
	if config.Global.OutputFormat != "json" {
		t.Error("OutputFormat должен быть json")
	}
	if config.Global.Verbosity != "verbose" {
		t.Error("Verbosity должен быть verbose")
	}
	if config.Global.MaxErrorsPerFile != 20 {
		t.Error("MaxErrorsPerFile должен быть 20")
	}
	
	if config.Autofix.Enabled {
		t.Error("Autofix должен быть отключен")
	}
}

// Тест загрузки несуществующего файла
func TestLoadConfigNotFound(t *testing.T) {
	config, err := LoadConfig("nonexistent.yaml")
	if err != nil {
		t.Errorf("Не должно быть ошибки при загрузке несуществующего файла: %v", err)
	}
	if config == nil {
		t.Error("Должна быть возвращена конфигурация по умолчанию")
	}
}

// Тест валидации конфигурации
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "Валидная конфигурация",
			config: &Config{
				Global: GlobalConfig{
					OutputFormat:     "text",
					Verbosity:        "normal",
					MaxErrorsPerFile: 10,
				},
			},
			expectError: false,
		},
		{
			name: "Невалидный формат вывода",
			config: &Config{
				Global: GlobalConfig{
					OutputFormat:     "invalid",
					Verbosity:        "normal",
					MaxErrorsPerFile: 10,
				},
			},
			expectError: true,
			errorMsg:    "недопустимый формат вывода",
		},
		{
			name: "Невалидный уровень детализации",
			config: &Config{
				Global: GlobalConfig{
					OutputFormat:     "text",
					Verbosity:        "invalid",
					MaxErrorsPerFile: 10,
				},
			},
			expectError: true,
			errorMsg:    "недопустимый уровень детализации",
		},
		{
			name: "Отрицательное количество ошибок",
			config: &Config{
				Global: GlobalConfig{
					OutputFormat:     "text",
					Verbosity:        "normal",
					MaxErrorsPerFile: -1,
				},
			},
			expectError: true,
			errorMsg:    "максимальное количество ошибок не может быть отрицательным",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.expectError {
				if err == nil {
					t.Error("Ожидалась ошибка, но ее не было")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Ожидалось сообщение об ошибке содержащее %q, получено: %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Не ожидалась ошибка, но получили: %v", err)
				}
			}
		})
	}
}

// Тест сохранения конфигурации
func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "save_test.yaml")
	
	config := DefaultConfig()
	config.Rules.LowercaseStart = false
	config.Global.OutputFormat = "json"
	
	err := SaveConfig(config, configPath)
	if err != nil {
		t.Fatalf("Ошибка сохранения конфигурации: %v", err)
	}
	
	// Проверяем, что файл создан
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Конфигурационный файл не был создан")
	}
	
	// Загружаем и проверяем
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Ошибка загрузки сохраненной конфигурации: %v", err)
	}
	
	if loadedConfig.Rules.LowercaseStart {
		t.Error("LowercaseStart должен быть отключен в загруженной конфигурации")
	}
	if loadedConfig.Global.OutputFormat != "json" {
		t.Error("OutputFormat должен быть json в загруженной конфигурации")
	}
}

// Тест функции contains
func TestContains(t *testing.T) {
	tests := []struct {
		slice []string
		item  string
		want  bool
	}{
		{
			slice: []string{"a", "b", "c"},
			item:  "b",
			want:  true,
		},
		{
			slice: []string{"a", "b", "c"},
			item:  "d",
			want:  false,
		},
		{
			slice: []string{},
			item:  "a",
			want:  false,
		},
		{
			slice: []string{"test"},
			item:  "test",
			want:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := contains(tt.slice, tt.item)
			if got != tt.want {
				t.Errorf("contains(%v, %q) = %v; want %v", tt.slice, tt.item, got, tt.want)
			}
		})
	}
}
