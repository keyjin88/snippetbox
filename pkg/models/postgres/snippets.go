package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"vav-tech.ru/snippetbox/pkg/models"
)

type SnippetModel struct {
	DB *pgxpool.Pool
}

// Insert - Метод для создания новой заметки в базе дынных.
func (s *SnippetModel) Insert(title, content string, expires int) (int, error) {
	stmt := `INSERT INTO snippets (title, content, created, expires) 
					VALUES($1, $2, now(), now() + interval '1 day' * $3) RETURNING id`
	var id int
	err := s.DB.QueryRow(context.Background(), stmt, title, content, expires).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, models.ErrorNoRecord
		} else {
			return 0, err
		}
	}
	return id, nil
}

// Get - Метод для возвращения данных заметки по её идентификатору ID.
func (s *SnippetModel) Get(id int) (*models.Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
    				WHERE expires > now() AND id = $1`
	row := s.DB.QueryRow(context.Background(), stmt, id)

	snippet := &models.Snippet{}
	err := row.Scan(
		&snippet.ID,
		&snippet.Title,
		&snippet.Content,
		&snippet.Created,
		&snippet.Expires,
	)
	if err != nil {
		// Специально для этого случая, мы проверим при помощи функции errors.Is()
		// если запрос был выполнен с ошибкой. Если ошибка обнаружена, то
		// возвращаем нашу ошибку из модели models.ErrNoRecord.
		if err.Error() == "no rows in result set" {
			return nil, models.ErrorNoRecord
		} else {
			return nil, err
		}
	}
	return snippet, nil
}

// Latest - Метод возвращает 10 наиболее часто используемые заметки.
func (s *SnippetModel) Latest() ([]*models.Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > now() ORDER BY created DESC LIMIT 10`
	rows, err := s.DB.Query(context.Background(), stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []*models.Snippet
	for rows.Next() {
		snippet := &models.Snippet{}
		err := rows.Scan(&snippet.ID, &snippet.Title, &snippet.Content, &snippet.Created, &snippet.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, snippet)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
