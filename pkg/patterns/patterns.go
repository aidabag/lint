package patterns

import (
	"fmt"
	"regexp"
	"strings"
)

// Менеджер паттернов для проверки чувствительных данных
type PatternManager struct {
	patterns     []*Pattern
	regexCache   map[string]*regexp.Regexp
}

// Структура паттерна
type Pattern struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type"`        // "simple", "regex", "custom"
	Pattern     string   `yaml:"pattern"`     // Паттерн для поиска
	Regex       string   `yaml:"regex"`       // Регулярное выражение
	Words       []string `yaml:"words"`       // Список слов для простого поиска
	Severity    string   `yaml:"severity"`    // "low", "medium", "high", "critical"
	Enabled     bool     `yaml:"enabled"`     // Включен ли паттерн
	Tags        []string `yaml:"tags"`        // Теги для группировки
}

// Создание нового менеджера паттернов
func NewPatternManager() *PatternManager {
	return &PatternManager{
		patterns:   make([]*Pattern, 0),
		regexCache: make(map[string]*regexp.Regexp),
	}
}

// Добавление паттерна
func (pm *PatternManager) AddPattern(pattern *Pattern) error {
	// Валидация паттерна
	if err := pm.validatePattern(pattern); err != nil {
		return fmt.Errorf("ошибка валидации паттерна %s: %w", pattern.Name, err)
	}
	
	// Компиляция регулярного выражения если нужно
	if pattern.Type == "regex" && pattern.Regex != "" {
		if _, exists := pm.regexCache[pattern.Regex]; !exists {
			regex, err := regexp.Compile(strings.ToLower(pattern.Regex))
			if err != nil {
				return fmt.Errorf("ошибка компиляции регулярного выражения %s: %w", pattern.Regex, err)
			}
			pm.regexCache[pattern.Regex] = regex
		}
	}
	
	pm.patterns = append(pm.patterns, pattern)
	return nil
}

// Проверка сообщения на наличие чувствительных данных
func (pm *PatternManager) CheckMessage(msg string) []*Match {
	matches := make([]*Match, 0)
	lowerMsg := strings.ToLower(msg)
	
	for _, pattern := range pm.patterns {
		if !pattern.Enabled {
			continue
		}
		
		if match := pm.checkPattern(pattern, lowerMsg, msg); match != nil {
			matches = append(matches, match)
		}
	}
	
	return matches
}

// Проверка конкретного паттерна
func (pm *PatternManager) checkPattern(pattern *Pattern, lowerMsg, originalMsg string) *Match {
	switch pattern.Type {
	case "simple":
		return pm.checkSimplePattern(pattern, lowerMsg, originalMsg)
	case "regex":
		return pm.checkRegexPattern(pattern, lowerMsg, originalMsg)
	case "custom":
		return pm.checkCustomPattern(pattern, lowerMsg, originalMsg)
	default:
		return nil
	}
}

// Проверка простого паттерна (поиск слов)
func (pm *PatternManager) checkSimplePattern(pattern *Pattern, lowerMsg, originalMsg string) *Match {
	for _, word := range pattern.Words {
		if strings.Contains(lowerMsg, word) {
			return &Match{
				Pattern:     pattern,
				MatchedText: word,
				Position:    strings.Index(lowerMsg, word),
				Severity:    pattern.Severity,
			}
		}
	}
	return nil
}

// Проверка паттерна с регулярным выражением
func (pm *PatternManager) checkRegexPattern(pattern *Pattern, lowerMsg, originalMsg string) *Match {
	regex, exists := pm.regexCache[pattern.Regex]
	if !exists {
		return nil
	}
	
	matches := regex.FindStringSubmatch(lowerMsg)
	if len(matches) > 0 {
		return &Match{
			Pattern:     pattern,
			MatchedText: matches[0],
			Position:    strings.Index(lowerMsg, matches[0]),
			Severity:    pattern.Severity,
		}
	}
	return nil
}

// Проверка кастомного паттерна
func (pm *PatternManager) checkCustomPattern(pattern *Pattern, lowerMsg, originalMsg string) *Match {
	// Здесь можно добавить сложную логику проверки
	// Например, проверка на форматы данных, номера карт, etc.
	
	switch pattern.Name {
	case "credit_card":
		return pm.checkCreditCard(lowerMsg, originalMsg)
	case "email":
		return pm.checkEmail(lowerMsg, originalMsg)
	case "phone":
		return pm.checkPhone(lowerMsg, originalMsg)
	case "ip_address":
		return pm.checkIPAddress(lowerMsg, originalMsg)
	case "url":
		return pm.checkURL(lowerMsg, originalMsg)
	default:
		return nil
	}
}

// Проверка номера кредитной карты
func (pm *PatternManager) checkCreditCard(lowerMsg, originalMsg string) *Match {
	// Простая проверка на последовательность цифр длиной 13-19
	cardRegex := regexp.MustCompile(`\b\d{13,19}\b`)
	if matches := cardRegex.FindStringSubmatch(lowerMsg); len(matches) > 0 {
		return &Match{
			Pattern:     &Pattern{Name: "credit_card", Type: "custom", Severity: "critical"},
			MatchedText: matches[0],
			Position:    strings.Index(lowerMsg, matches[0]),
			Severity:    "critical",
		}
	}
	return nil
}

// Проверка email
func (pm *PatternManager) checkEmail(lowerMsg, originalMsg string) *Match {
	emailRegex := regexp.MustCompile(`\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b`)
	if matches := emailRegex.FindStringSubmatch(lowerMsg); len(matches) > 0 {
		return &Match{
			Pattern:     &Pattern{Name: "email", Type: "custom", Severity: "medium"},
			MatchedText: matches[0],
			Position:    strings.Index(lowerMsg, matches[0]),
			Severity:    "medium",
		}
	}
	return nil
}

// Проверка номера телефона
func (pm *PatternManager) checkPhone(lowerMsg, originalMsg string) *Match {
	phoneRegex := regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b|\+?\d{1,3}[-.]?\d{3}[-.]?\d{3}[-.]?\d{4}\b`)
	if matches := phoneRegex.FindStringSubmatch(lowerMsg); len(matches) > 0 {
		return &Match{
			Pattern:     &Pattern{Name: "phone", Type: "custom", Severity: "medium"},
			MatchedText: matches[0],
			Position:    strings.Index(lowerMsg, matches[0]),
			Severity:    "medium",
		}
	}
	return nil
}

// Проверка IP адреса
func (pm *PatternManager) checkIPAddress(lowerMsg, originalMsg string) *Match {
	ipRegex := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	if matches := ipRegex.FindStringSubmatch(lowerMsg); len(matches) > 0 {
		return &Match{
			Pattern:     &Pattern{Name: "ip_address", Type: "custom", Severity: "low"},
			MatchedText: matches[0],
			Position:    strings.Index(lowerMsg, matches[0]),
			Severity:    "low",
		}
	}
	return nil
}

// Проверка URL
func (pm *PatternManager) checkURL(lowerMsg, originalMsg string) *Match {
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	if matches := urlRegex.FindStringSubmatch(lowerMsg); len(matches) > 0 {
		return &Match{
			Pattern:     &Pattern{Name: "url", Type: "custom", Severity: "low"},
			MatchedText: matches[0],
			Position:    strings.Index(lowerMsg, matches[0]),
			Severity:    "low",
		}
	}
	return nil
}

// Результат совпадения
type Match struct {
	Pattern     *Pattern `json:"pattern"`
	MatchedText string   `json:"matched_text"`
	Position    int      `json:"position"`
	Severity    string   `json:"severity"`
}

// Валидация паттерна
func (pm *PatternManager) validatePattern(pattern *Pattern) error {
	if pattern.Name == "" {
		return fmt.Errorf("имя паттерна не может быть пустым")
	}
	
	if pattern.Type == "" {
		return fmt.Errorf("тип паттерна не может быть пустым")
	}
	
	switch pattern.Type {
	case "simple":
		if len(pattern.Words) == 0 {
			return fmt.Errorf("для простого паттерна нужны слова для поиска")
		}
	case "regex":
		if pattern.Regex == "" {
			return fmt.Errorf("для regex паттерна нужно регулярное выражение")
		}
		// Проверка валидности регулярного выражения
		_, err := regexp.Compile(pattern.Regex)
		if err != nil {
			return fmt.Errorf("невалидное регулярное выражение: %w", err)
		}
	case "custom":
		// Кастомные паттерны предопределены
		validCustomPatterns := []string{"credit_card", "email", "phone", "ip_address", "url"}
		found := false
		for _, valid := range validCustomPatterns {
			if pattern.Name == valid {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("неизвестный кастомный паттерн: %s", pattern.Name)
		}
	default:
		return fmt.Errorf("неизвестный тип паттерна: %s", pattern.Type)
	}
	
	if pattern.Severity == "" {
		pattern.Severity = "medium"
	}
	
	return nil
}

// Получение всех паттернов
func (pm *PatternManager) GetPatterns() []*Pattern {
	return pm.patterns
}

// Получение паттернов по тегу
func (pm *PatternManager) GetPatternsByTag(tag string) []*Pattern {
	result := make([]*Pattern, 0)
	for _, pattern := range pm.patterns {
		for _, patternTag := range pattern.Tags {
			if patternTag == tag {
				result = append(result, pattern)
				break
			}
		}
	}
	return result
}

// Включение/выключение паттерна по имени
func (pm *PatternManager) SetPatternEnabled(name string, enabled bool) bool {
	for _, pattern := range pm.patterns {
		if pattern.Name == name {
			pattern.Enabled = enabled
			return true
		}
	}
	return false
}

// Создание паттернов по умолчанию
func (pm *PatternManager) LoadDefaultPatterns() {
	defaultPatterns := []*Pattern{
		{
			Name:        "password",
			Description: "Пароли и связанные с ними данные",
			Type:        "simple",
			Words:       []string{"password", "passwd", "pwd", "pass"},
			Severity:    "critical",
			Enabled:     true,
			Tags:        []string{"auth", "credentials"},
		},
		{
			Name:        "api_key",
			Description: "API ключи",
			Type:        "simple",
			Words:       []string{"api_key", "apikey", "api-key", "access_key", "access-key"},
			Severity:    "critical",
			Enabled:     true,
			Tags:        []string{"auth", "api"},
		},
		{
			Name:        "token",
			Description: "Токены аутентификации",
			Type:        "simple",
			Words:       []string{"token", "jwt", "bearer", "oauth"},
			Severity:    "high",
			Enabled:     true,
			Tags:        []string{"auth"},
		},
		{
			Name:        "secret",
			Description: "Секретные данные",
			Type:        "simple",
			Words:       []string{"secret", "private", "confidential"},
			Severity:    "high",
			Enabled:     true,
			Tags:        []string{"security"},
		},
		{
			Name:        "credit_card",
			Description: "Номера кредитных карт",
			Type:        "custom",
			Severity:    "critical",
			Enabled:     true,
			Tags:        []string{"financial", "pii"},
		},
		{
			Name:        "email",
			Description: "Email адреса",
			Type:        "custom",
			Severity:    "medium",
			Enabled:     true,
			Tags:        []string{"pii", "contact"},
		},
		{
			Name:        "phone",
			Description: "Номера телефонов",
			Type:        "custom",
			Severity:    "medium",
			Enabled:     true,
			Tags:        []string{"pii", "contact"},
		},
		{
			Name:        "ip_address",
			Description: "IP адреса",
			Type:        "custom",
			Severity:    "low",
			Enabled:     true,
			Tags:        []string{"network"},
		},
		{
			Name:        "url",
			Description: "URL адреса",
			Type:        "custom",
			Severity:    "low",
			Enabled:     true,
			Tags:        []string{"network"},
		},
		{
			Name:        "base64",
			Description: "Base64 закодированные данные (длинные последовательности)",
			Type:        "regex",
			Regex:       `[A-Za-z0-9+/]{40,}={0,2}`,
			Severity:    "medium",
			Enabled:     true,
			Tags:        []string{"encoding"},
		},
	}
	
	for _, pattern := range defaultPatterns {
		pm.AddPattern(pattern)
	}
}
