package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitService предоставляет методы для работы с Git репозиторием
type GitService struct {
	root string
}

// NewGitService создаёт новый сервис для работы с Git
func NewGitService(repoPath string) (*GitService, error) {
	// Проверяем, что это Git репозиторий
	if !IsGitRepository(repoPath) {
		return nil, fmt.Errorf("директория %s не является Git репозиторием", repoPath)
	}

	return &GitService{
		root: repoPath,
	}, nil
}

// GetLastCommitMessages возвращает последние сообщения коммитов для файла
func (gs *GitService) GetLastCommitMessages(filePath string, n int) ([]string, error) {
	// Получаем относительный путь файла от корня репозитория
	relPath, err := filepath.Rel(gs.root, filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения относительного пути: %w", err)
	}

	// Нормализуем разделители путей для Git
	relPath = strings.ReplaceAll(relPath, "\\", "/")

	// Выполняем git log команду
	cmd := exec.Command("git", "log", "--oneline", "-n", fmt.Sprintf("%d", n), "--", relPath)
	cmd.Dir = gs.root

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения git log: %w", err)
	}

	// Парсим вывод
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var messages []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Убираем хеш коммита (первые 7 символов + пробел)
			if len(line) > 8 {
				message := strings.TrimSpace(line[8:])
				if message != "" {
					messages = append(messages, message)
				}
			}
		}
	}

	return messages, nil
}

// IsGitRepository проверяет, является ли директория Git репозиторием
func IsGitRepository(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path

	err := cmd.Run()
	return err == nil
}
