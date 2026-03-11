package patterns

import (
	"testing"
)

// Тест создания менеджера паттернов
func TestNewPatternManager(t *testing.T) {
	pm := NewPatternManager()
	if pm == nil {
		t.Error("Менеджер паттернов не должен быть nil")
	}
	if len(pm.patterns) != 0 {
		t.Error("Новый менеджер должен быть пустым")
	}
	if len(pm.regexCache) != 0 {
		t.Error("Кэш регулярных выражений должен быть пустым")
	}
}

// Тест добавления простого паттерна
func TestAddSimplePattern(t *testing.T) {
	pm := NewPatternManager()
	
	pattern := &Pattern{
		Name:     "test",
		Type:     "simple",
		Words:    []string{"test", "example"},
		Severity: "medium",
		Enabled:  true,
	}
	
	err := pm.AddPattern(pattern)
	if err != nil {
		t.Fatalf("Ошибка добавления паттерна: %v", err)
	}
	
	if len(pm.patterns) != 1 {
		t.Error("Должен быть один паттерн")
	}
	
	if pm.patterns[0].Name != "test" {
		t.Error("Имя паттерна должно быть 'test'")
	}
}

// Тест добавления regex паттерна
func TestAddRegexPattern(t *testing.T) {
	pm := NewPatternManager()
	
	pattern := &Pattern{
		Name:     "regex_test",
		Type:     "regex",
		Regex:    `\d{4}-\d{2}-\d{2}`,
		Severity: "low",
		Enabled:  true,
	}
	
	err := pm.AddPattern(pattern)
	if err != nil {
		t.Fatalf("Ошибка добавления regex паттерна: %v", err)
	}
	
	if len(pm.regexCache) != 1 {
		t.Error("В кэше должно быть одно регулярное выражение")
	}
}

// Тест валидации паттернов
func TestValidatePattern(t *testing.T) {
	pm := NewPatternManager()
	
	tests := []struct {
		name        string
		pattern     *Pattern
		expectError bool
		errorMsg    string
	}{
		{
			name: "Валидный простой паттерн",
			pattern: &Pattern{
				Name:     "valid",
				Type:     "simple",
				Words:    []string{"test"},
				Severity: "medium",
			},
			expectError: false,
		},
		{
			name: "Пустое имя",
			pattern: &Pattern{
				Type:  "simple",
				Words: []string{"test"},
			},
			expectError: true,
			errorMsg:    "имя паттерна не может быть пустым",
		},
		{
			name: "Пустой тип",
			pattern: &Pattern{
				Name:  "test",
				Words: []string{"test"},
			},
			expectError: true,
			errorMsg:    "тип паттерна не может быть пустым",
		},
		{
			name: "Простой паттерн без слов",
			pattern: &Pattern{
				Name:     "test",
				Type:     "simple",
				Severity: "medium",
			},
			expectError: true,
			errorMsg:    "для простого паттерна нужны слова для поиска",
		},
		{
			name: "Regex паттерн без выражения",
			pattern: &Pattern{
				Name:     "test",
				Type:     "regex",
				Severity: "medium",
			},
			expectError: true,
			errorMsg:    "для regex паттерна нужно регулярное выражение",
		},
		{
			name: "Невалидное регулярное выражение",
			pattern: &Pattern{
				Name:     "test",
				Type:     "regex",
				Regex:    "[invalid",
				Severity: "medium",
			},
			expectError: true,
			errorMsg:    "невалидное регулярное выражение",
		},
		{
			name: "Неизвестный тип паттерна",
			pattern: &Pattern{
				Name:     "test",
				Type:     "unknown",
				Severity: "medium",
			},
			expectError: true,
			errorMsg:    "неизвестный тип паттерна",
		},
		{
			name: "Неизвестный кастомный паттерн",
			pattern: &Pattern{
				Name: "unknown_custom",
				Type: "custom",
			},
			expectError: true,
			errorMsg:    "неизвестный кастомный паттерн",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.validatePattern(tt.pattern)
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

// Тест проверки сообщений
func TestCheckMessage(t *testing.T) {
	pm := NewPatternManager()
	
	// Добавляем тестовые паттерны
	patterns := []*Pattern{
		{
			Name:     "password",
			Type:     "simple",
			Words:    []string{"password", "pwd"},
			Severity: "critical",
			Enabled:  true,
		},
		{
			Name:     "email",
			Type:     "regex",
			Regex:    `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
			Severity: "medium",
			Enabled:  true,
		},
	}
	
	for _, pattern := range patterns {
		pm.AddPattern(pattern)
	}
	
	tests := []struct {
		name         string
		message      string
		expectMatches int
	}{
		{
			name:         "Сообщение с паролем",
			message:      "user password: secret123",
			expectMatches: 1,
		},
		{
			name:         "Сообщение с email",
			message:      "user email: test@example.com",
			expectMatches: 1,
		},
		{
			name:         "Сообщение с паролем и email",
			message:      "user password: secret123, email: test@example.com",
			expectMatches: 2,
		},
		{
			name:         "Чистое сообщение",
			message:      "user logged in successfully",
			expectMatches: 0,
		},
		{
			name:         "Сообщение с отключенным паттерном",
			message:      "user password: secret123",
			expectMatches: 0,
		},
	}
	
	// Тест с включенными паттернами
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Сообщение с отключенным паттерном" {
				// Отключаем паттерн password
				pm.SetPatternEnabled("password", false)
			}
			
			matches := pm.CheckMessage(tt.message)
			if len(matches) != tt.expectMatches {
				t.Errorf("Ожидалось %d совпадений, получено %d для сообщения %q", tt.expectMatches, len(matches), tt.message)
			}
			
			// Возвращаем паттерн обратно
			if tt.name == "Сообщение с отключенным паттерном" {
				pm.SetPatternEnabled("password", true)
			}
		})
	}
}

// Тест кастомных паттернов
func TestCustomPatterns(t *testing.T) {
	pm := NewPatternManager()
	
	// Добавляем кастомные паттерны
	customPatterns := []*Pattern{
		{Name: "credit_card", Type: "custom", Enabled: true},
		{Name: "email", Type: "custom", Enabled: true},
		{Name: "phone", Type: "custom", Enabled: true},
		{Name: "ip_address", Type: "custom", Enabled: true},
		{Name: "url", Type: "custom", Enabled: true},
	}
	
	for _, pattern := range customPatterns {
		err := pm.AddPattern(pattern)
		if err != nil {
			t.Fatalf("Ошибка добавления кастомного паттерна %s: %v", pattern.Name, err)
		}
	}
	
	tests := []struct {
		name         string
		message      string
		expectedType string
		expectMatch  bool
	}{
		{
			name:         "Кредитная карта",
			message:      "card number: 4111111111111111",
			expectedType: "credit_card",
			expectMatch:  true,
		},
		{
			name:         "Email",
			message:      "contact: user@example.com",
			expectedType: "email",
			expectMatch:  true,
		},
		{
			name:         "Телефон",
			message:      "phone: 123-456-7890",
			expectedType: "phone",
			expectMatch:  true,
		},
		{
			name:         "IP адрес",
			message:      "server: 192.168.1.1",
			expectedType: "ip_address",
			expectMatch:  true,
		},
		{
			name:         "URL",
			message:      "link: https://example.com",
			expectedType: "url",
			expectMatch:  true,
		},
		{
			name:         "Не совпадает",
			message:      "simple log message",
			expectedType: "",
			expectMatch:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := pm.CheckMessage(tt.message)
			
			found := false
			for _, match := range matches {
				if match.Pattern.Name == tt.expectedType {
					found = true
					break
				}
			}
			
			if tt.expectMatch && !found {
				t.Errorf("Ожидалось совпадение для типа %s в сообщении %q", tt.expectedType, tt.message)
			}
			if !tt.expectMatch && found {
				t.Errorf("Не ожидалось совпадение для сообщения %q", tt.message)
			}
		})
	}
}

// Тест паттернов по умолчанию
func TestLoadDefaultPatterns(t *testing.T) {
	pm := NewPatternManager()
	pm.LoadDefaultPatterns()
	
	patterns := pm.GetPatterns()
	if len(patterns) == 0 {
		t.Error("Должны быть загружены паттерны по умолчанию")
	}
	
	// Проверяем наличие основных паттернов
	expectedPatterns := []string{"password", "api_key", "token", "secret", "credit_card", "email"}
	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range patterns {
			if pattern.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Должен быть паттерн %s", expected)
		}
	}
}

// Тест получения паттернов по тегу
func TestGetPatternsByTag(t *testing.T) {
	pm := NewPatternManager()
	pm.LoadDefaultPatterns()
	
	authPatterns := pm.GetPatternsByTag("auth")
	if len(authPatterns) == 0 {
		t.Error("Должны быть паттерны с тегом 'auth'")
	}
	
	// Проверяем, что все паттерны имеют тег 'auth'
	for _, pattern := range authPatterns {
		found := false
		for _, tag := range pattern.Tags {
			if tag == "auth" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Паттерн %s должен иметь тег 'auth'", pattern.Name)
		}
	}
}

// Вспомогательная функция
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   (len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		   (len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
