package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/aidabag/lint/pkg/config"
)

// Создание простой версии линтера без использования x/tools
func main() {
	// Парсинг флагов командной строки
	var (
		configPath = flag.String("config", ".loglinter.yaml", "Путь к конфигурационному файлу")
		autofix    = flag.Bool("fix", false, "Автоматически исправлять ошибки")
		verbose    = flag.Bool("v", false, "Подробный вывод")
	)
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Использование: simple [опции] <путь>")
		fmt.Println("Опции:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Загрузка конфигурации
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Используется конфигурация из файла: %s\n", *configPath)
	}

	path := flag.Args()[0]
	err = processPath(path, cfg, *autofix, *verbose)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		os.Exit(1)
	}
}

// Обработка пути к файлам или директории
func processPath(path string, cfg *config.Config, autofix, verbose bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("ошибка доступа к пути %s: %w", path, err)
	}

	if info.IsDir() {
		return processDirectory(path, cfg, autofix, verbose)
	}
	return processFile(path, cfg, autofix, verbose)
}

// Обработка директории
func processDirectory(dir string, cfg *config.Config, autofix, verbose bool) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Проверяем исключения
		if shouldExclude(path, cfg) {
			if verbose {
				fmt.Printf("Исключен файл: %s\n", path)
			}
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			if err := processFile(path, cfg, autofix, verbose); err != nil {
				fmt.Printf("Ошибка в файле %s: %v\n", path, err)
			}
		}
		return nil
	})
}

// Проверка, должен ли файл быть исключен
func shouldExclude(path string, cfg *config.Config) bool {
	for _, pattern := range cfg.Global.Exclude.Patterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		// Проверяем путь относительно текущей директории
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}
	return false
}

// Обработка одного файла
func processFile(filePath string, cfg *config.Config, autofix, verbose bool) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("ошибка парсинга файла %s: %w", filePath, err)
	}

	if verbose {
		fmt.Printf("Проверка файла: %s\n", filePath)
	}
	
	// Создание простого анализатора
	analyzer := &simpleAnalyzer{
		fset:    fset,
		file:    filePath,
		config:  cfg,
		verbose: verbose,
	}
	
	analyzer.analyze(node)
	
	// Авто-исправление
	if autofix && analyzer.found && analyzer.hasFixes {
		if err := analyzer.applyFixes(); err != nil {
			return fmt.Errorf("ошибка применения исправлений: %w", err)
		}
		if verbose {
			fmt.Printf("  Применены авто-исправления для файла: %s\n", filePath)
		}
	}
	
	return nil
}

// Простой анализатор
type simpleAnalyzer struct {
	fset     *token.FileSet
	file     string
	config   *config.Config
	verbose  bool
	found    bool
	hasFixes bool
	fixes    []fix
}

// Структура для исправления
type fix struct {
	line    int
	oldText string
	newText string
	rule    string
}

// Анализ AST
func (a *simpleAnalyzer) analyze(node any) {
	// Здесь можно добавить базовый анализ
	// но для демонстрации просто проверим наличие лог-вызовов в файле
	
	content, err := os.ReadFile(a.file)
	if err != nil {
		return
	}
	
	lines := strings.Split(string(content), "\n")
	errorCount := 0
	
	for i, line := range lines {
		if errorCount >= a.config.Global.MaxErrorsPerFile {
			if a.verbose {
				fmt.Printf("  Достигнуто максимальное количество ошибок для файла: %d\n", a.config.Global.MaxErrorsPerFile)
			}
			break
		}
		
		if a.isLogCall(line) {
			// Простая проверка на наличие потенциальных проблем
			if strings.Contains(line, `"`) {
				start := strings.Index(line, `"`)
				if start >= 0 {
					end := strings.Index(line[start+1:], `"`)
					if end >= 0 {
						msg := line[start+1 : start+1+end]
						if a.checkMessage(msg, i+1, line) {
							errorCount++
						}
					}
				}
			}
		}
	}
}

// Проверка, что это вызов лог-функции
func (a *simpleAnalyzer) isLogCall(line string) bool {
	// Проверяем поддерживаемые логгеры
	for _, logger := range a.config.Loggers.Supported {
		if strings.Contains(line, logger+".") {
			// Проверяем, что это не исключенная функция
			for _, excluded := range a.config.Loggers.ExcludedFunctions {
				if strings.Contains(line, logger+"."+excluded) {
					return false
				}
			}
			return true
		}
	}
	return false
}

// Проверка сообщения
func (a *simpleAnalyzer) checkMessage(msg string, lineNum int, fullLine string) bool {
	foundIssue := false
	
	// Правило 1: строчная буква в начале
	if a.config.Rules.LowercaseStart && len(msg) > 0 && msg[0] >= 'A' && msg[0] <= 'Z' && !a.isProbablyName(msg) {
		fmt.Printf("  Строка %d: лог-сообщение должно начинаться со строчной буквы: %q\n", lineNum, msg)
		a.found = true
		foundIssue = true
		
		// Добавляем исправление
		if a.config.Autofix.Enabled && a.contains(a.config.Autofix.Rules, "lowercase_start") {
			fixedMsg := strings.ToLower(msg[:1]) + msg[1:]
			a.addFix(lineNum, msg, fixedMsg, "lowercase_start")
		}
	}
	
	// Правило 2: английский язык
	if a.config.Rules.EnglishOnly && !a.isEnglish(msg) {
		fmt.Printf("  Строка %d: лог-сообщение должно быть на английском языке: %q\n", lineNum, msg)
		a.found = true
		foundIssue = true
	}
	
	// Правило 3: спецсимволы
	if a.config.Rules.NoSpecialChars && a.hasSpecialChars(msg) && !a.isProbablyFormatString(msg) {
		fmt.Printf("  Строка %d: лог-сообщение не должно содержать спецсимволы или эмодзи: %q\n", lineNum, msg)
		a.found = true
		foundIssue = true
	}
	
	// Правило 4: чувствительные данные
	if a.config.Rules.NoSensitiveData && a.hasSensitiveData(msg) && !a.isProbablySafeContext(msg) {
		fmt.Printf("  Строка %d: лог-сообщение не должно содержать чувствительные данные: %q\n", lineNum, msg)
		a.found = true
		foundIssue = true
	}
	
	return foundIssue
}

// Добавление исправления
func (a *simpleAnalyzer) addFix(line int, oldText, newText, rule string) {
	a.fixes = append(a.fixes, fix{
		line:    line,
		oldText: oldText,
		newText: newText,
		rule:    rule,
	})
	a.hasFixes = true
}

// Применение исправлений
func (a *simpleAnalyzer) applyFixes() error {
	if !a.hasFixes {
		return nil
	}
	
	content, err := os.ReadFile(a.file)
	if err != nil {
		return err
	}
	
	lines := strings.Split(string(content), "\n")
	
	// Создаем backup если нужно
	if a.config.Autofix.Backup {
		backupPath := a.file + ".backup"
		if err := os.WriteFile(backupPath, content, 0644); err != nil {
			return fmt.Errorf("ошибка создания backup файла: %w", err)
		}
	}
	
	// Применяем исправления в обратном порядке, чтобы не сбить номера строк
	for i := len(a.fixes) - 1; i >= 0; i-- {
		fix := a.fixes[i]
		if fix.line-1 < len(lines) {
			lines[fix.line-1] = strings.Replace(lines[fix.line-1], fix.oldText, fix.newText, 1)
		}
	}
	
	// Записываем исправленный файл
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(a.file, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("ошибка записи исправленного файла: %w", err)
	}
	
	return nil
}

// Проверка, что сообщение может быть именем (допускает заглавную букву)
func (a *simpleAnalyzer) isProbablyName(msg string) bool {
	// Короткие сообщения (1-2 слова) могут быть именами
	words := strings.Fields(msg)
	if len(words) <= 2 && len(msg) < a.config.Settings.LowercaseStart.MaxLengthException {
		return true
	}
	
	// Сообщения с proper nouns из конфигурации
	for _, noun := range a.config.Settings.LowercaseStart.AllowedWords {
		if strings.Contains(msg, noun) {
			return true
		}
	}
	
	return false
}

// Проверка на английский язык с учетом конфигурации
func (a *simpleAnalyzer) isEnglish(msg string) bool {
	for _, r := range msg {
		// Проверка на символы вне базового ASCII
		if r > 127 && !unicode.IsPunct(r) && !unicode.IsSpace(r) {
			// Проверяем, не является ли символ разрешенным
			charStr := string(r)
			allowed := false
			for _, allowedChar := range a.config.Settings.EnglishOnly.AllowedUnicode {
				if charStr == allowedChar {
					allowed = true
					break
				}
			}
			if !allowed {
				return false
			}
		}
	}
	return true
}

// Проверка на спецсимволы с учетом конфигурации
func (a *simpleAnalyzer) hasSpecialChars(msg string) bool {
	for _, r := range msg {
		// Проверка на эмодзи
		if a.config.Settings.NoSpecialChars.BlockEmoji {
			if r >= 0x1F600 && r <= 0x1F64F { // Эмотиконы
				return true
			}
			if r >= 0x1F300 && r <= 0x1F5FF { // Символы и пиктограммы
				return true
			}
			if r >= 0x1F680 && r <= 0x1F6FF { // Транспорт и символы карт
				return true
			}
			if r >= 0x1F700 && r <= 0x1F77F { // Алхимические символы
				return true
			}
			if r >= 0x1F780 && r <= 0x1F7FF { // Геометрические символы
				return true
			}
			if r >= 0x1F800 && r <= 0x1F8FF { // Дополнительные стрелки
				return true
			}
			if r >= 0x1F900 && r <= 0x1F9FF { // Дополнительные символы
				return true
			}
			if r >= 0x1FA00 && r <= 0x1FA6F { // Символы шахмат
				return true
			}
			if r >= 0x1FA70 && r <= 0x1FAFF { // Символы в брайле
				return true
			}
		}
		
		// Проверка на избыточную пунктуацию
		if a.config.Settings.NoSpecialChars.BlockRepeatedPunctuation {
			if strings.ContainsRune("!@#$%^&*()[]{}|\\:;\"'<>?,./", r) {
				return true
			}
		}
	}
	return false
}

// Проверка на чувствительные данные с учетом конфигурации
func (a *simpleAnalyzer) hasSensitiveData(msg string) bool {
	lowerMsg := strings.ToLower(msg)
	
	// Проверяем паттерны из конфигурации
	for _, pattern := range a.config.Settings.NoSensitiveData.SensitivePatterns {
		if strings.Contains(lowerMsg, pattern) {
			return true
		}
	}
	
	return false
}

// Проверка безопасного контекста с учетом конфигурации
func (a *simpleAnalyzer) isProbablySafeContext(msg string) bool {
	lowerMsg := strings.ToLower(msg)
	
	// Проверяем безопасные контексты из конфигурации
	for _, ctx := range a.config.Settings.NoSensitiveData.SafeContexts {
		if strings.Contains(lowerMsg, ctx) {
			return true
		}
	}
	
	return false
}

// Проверка, что сообщение может быть строкой форматирования
func (a *simpleAnalyzer) isProbablyFormatString(msg string) bool {
	// Строки форматирования с %s, %d, %v и т.д.
	formatSpecifiers := []string{"%s", "%d", "%v", "%f", "%t", "%q", "%x", "%+v", "%#v"}
	for _, spec := range formatSpecifiers {
		if strings.Contains(msg, spec) {
			return true
		}
	}
	
	// Строки с фигурными скобками для zap
	if strings.Contains(msg, "{") && strings.Contains(msg, "}") {
		return true
	}
	
	return false
}

// Проверка наличия элемента в срезе
func (a *simpleAnalyzer) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
