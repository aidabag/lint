package main

import "log/slog"

func testFunction() {
	// Примеры плохих логов
	slog.Info("Starting server")           // Должно начинаться со строчной буквы
	slog.Error("Ошибка подключения")       // Должно быть на английском
	slog.Warn("warning! Something wrong")  // Не должно содержать спецсимволы
	slog.Info("Password: secret123")       // Не должно содержать чувствительные данные
	slog.Debug("API key: abc123")          // Не должно содержать чувствительные данные
	slog.Info("server started! 🚀")        // Не должно содержать эмодзи
	
	// Примеры хороших логов
	slog.Info("starting server")
	slog.Error("failed to connect")
	slog.Warn("something went wrong")
	slog.Info("user authenticated successfully")
	slog.Debug("api request completed")
	slog.Info("token validated")
}
