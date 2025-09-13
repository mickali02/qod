// Filename: internal/data/comments.go
package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	// This is the validator from slides 165-168
	"github.com/mickali02/qod/internal/validator"
)

// This is the Comment struct from slide 176
type Comment struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"-"` // The "-" tag means this field will be hidden in JSON responses
	Version   int32     `json:"version"`
}

// This is the validation function from slide 173
func ValidateComment(v *validator.Validator, comment *Comment) {
	v.Check(comment.Content != "", "content", "must be provided")
	v.Check(len(comment.Content) <= 100, "content", "must not be more than 100 bytes long")

	v.Check(comment.Author != "", "author", "must be provided")
	v.Check(len(comment.Author) <= 25, "author", "must not be more than 25 bytes long")
}

// This is the CommentModel from slide 181
type CommentModel struct {
	DB *sql.DB
}

// This is the Insert method from slides 182-183
func (c CommentModel) Insert(comment *Comment) error {
	query := `
		 INSERT INTO comments (content, author)
		 VALUES ($1, $2)
		 RETURNING id, created_at, version`

	args := []any{comment.Content, comment.Author}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&comment.ID, &comment.CreatedAt, &comment.Version)
}

// This is the Get method from slides 191-193
func (c CommentModel) Get(id int64) (*Comment, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
		 SELECT id, created_at, content, author, version
		 FROM comments
		 WHERE id = $1`

	var comment Comment
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, id).Scan(&comment.ID, &comment.CreatedAt, &comment.Content, &comment.Author, &comment.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &comment, nil
}

// This is the Update method from slides 208-209
func (c CommentModel) Update(comment *Comment) error {
	query := `
		UPDATE comments
		SET content = $1, author = $2, version = version + 1
		WHERE id = $3
		RETURNING version`

	args := []any{comment.Content, comment.Author, comment.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&comment.Version)
}

// This is the Delete method from slides 220-222
func (c CommentModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
		DELETE FROM comments
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := c.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// This is the GetAll method from slides 248-250 (the version with filtering)
func (c CommentModel) GetAll(content string, author string) ([]*Comment, error) {
	query := `
		SELECT id, created_at, content, author, version
		FROM comments
		WHERE (to_tsvector('simple', content) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (to_tsvector('simple', author) @@ plainto_tsquery('simple', $2) OR $2 = '')
		ORDER BY id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, content, author)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []*Comment{}

	for rows.Next() {
		var comment Comment
		err := rows.Scan(
			&comment.ID,
			&comment.CreatedAt,
			&comment.Content,
			&comment.Author,
			&comment.Version,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

