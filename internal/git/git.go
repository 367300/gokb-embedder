package git

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitService предоставляет методы для работы с Git репозиторием
type GitService struct {
	repo *git.Repository
	root string
}

// NewGitService создаёт новый сервис для работы с Git
func NewGitService(repoPath string) (*GitService, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия репозитория: %w", err)
	}

	// Получаем корневую директорию репозитория
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения рабочего дерева: %w", err)
	}

	return &GitService{
		repo: repo,
		root: worktree.Filesystem.Root(),
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

	// Получаем историю коммитов для файла
	commits, err := gs.getFileCommits(relPath, n)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения коммитов для файла %s: %w", filePath, err)
	}

	// Извлекаем сообщения коммитов
	var messages []string
	for _, commit := range commits {
		messages = append(messages, strings.TrimSpace(commit.Message))
	}

	return messages, nil
}

// getFileCommits получает коммиты для конкретного файла
func (gs *GitService) getFileCommits(filePath string, n int) ([]*object.Commit, error) {
	// Получаем HEAD коммит
	head, err := gs.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения HEAD: %w", err)
	}

	// Получаем коммит
	commit, err := gs.repo.CommitObject(head.Hash)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения коммита: %w", err)
	}

	var commits []*object.Commit
	count := 0

	// Итерируемся по истории коммитов
	err = gs.iterateCommits(commit, func(c *object.Commit) error {
		if count >= n {
			return fmt.Errorf("достигнут лимит коммитов")
		}

		// Проверяем, изменялся ли файл в этом коммите
		if gs.fileChangedInCommit(c, filePath) {
			commits = append(commits, c)
			count++
		}

		return nil
	})

	if err != nil && err.Error() != "достигнут лимит коммитов" {
		return nil, fmt.Errorf("ошибка итерации коммитов: %w", err)
	}

	return commits, nil
}

// iterateCommits итерируется по истории коммитов
func (gs *GitService) iterateCommits(commit *object.Commit, fn func(*object.Commit) error) error {
	if commit == nil {
		return nil
	}

	if err := fn(commit); err != nil {
		return err
	}

	// Получаем родительские коммиты
	for _, parent := range commit.ParentHashes {
		parentCommit, err := gs.repo.CommitObject(parent)
		if err != nil {
			continue // Пропускаем недоступные коммиты
		}

		if err := gs.iterateCommits(parentCommit, fn); err != nil {
			return err
		}
	}

	return nil
}

// fileChangedInCommit проверяет, изменялся ли файл в коммите
func (gs *GitService) fileChangedInCommit(commit *object.Commit, filePath string) bool {
	// Получаем изменения в коммите
	changes, err := commit.Parent(0)
	if err != nil {
		// Если нет родителя (первый коммит), считаем что файл изменялся
		return true
	}

	patch, err := changes.Patch(commit)
	if err != nil {
		return false
	}

	// Проверяем, есть ли изменения в нашем файле
	for _, filePatch := range patch.FilePatches() {
		from, to := filePatch.Files()
		if (from != nil && from.Name() == filePath) || (to != nil && to.Name() == filePath) {
			return true
		}
	}

	return false
}

// IsGitRepository проверяет, является ли директория Git репозиторием
func IsGitRepository(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}
