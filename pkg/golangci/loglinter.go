package golangci

import (
	"github.com/aidabag/lint/pkg/loglinter"
	"golang.org/x/tools/go/analysis"
)

// Создание конфигурации для линтера
type Config struct {
	// Правила, которые нужно проверять
	Rules []string `mapstructure:"rules"`
}

// Создание нового линтера для golangci-lint
func NewLogLinter(cfg *Config) *analysis.Analyzer {
	analyzer := loglinter.Analyzer
	
	// Здесь можно добавить настройку правил на основе конфигурации
	// если это потребуется в будущем
	
	return analyzer
}

// Получение имени линтера
func GetLinterName() string {
	return "loglinter"
}

// Получение описания линтера
func GetLinterDescription() string {
	return "Проверяет лог-сообщения на соответствие правилам: строчная буква в начале, английский язык, отсутствие спецсимволов и чувствительных данных"
}
