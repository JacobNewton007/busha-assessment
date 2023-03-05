package data

import (
	"context"
	"database/sql"
	"time"
)

type Comment struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"-"`
	Comment     string    `json:"comment"`
	Movie       string    `json:"movie_name"`
	CommenterIp string    `json:"commenter_ip"`
	Version     int32     `json:"version"`
}

type CommentModels struct {
	DB *sql.DB
}

func (c CommentModels) Insert(comment *Comment) error {

	query := `
		INSERT INTO comments (comment, movie_name, commenter_ip)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, version`

	args := []interface{}{comment.Comment, comment.Movie, comment.CommenterIp}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&comment.ID, &comment.CreatedAt, &comment.Version)
}

func (c CommentModels) GetCommentForMovie(movie_name string) ([]*Comment,  int,  error) {

	query := `
		SELECT COUNT(*) OVER(), id, created_at, comment, movie_name, commenter_ip FROM comments WHERE movie_name = $1
		GROUP BY id, created_at, comment, movie_name, commenter_ip
		ORDER BY id DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, movie_name)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	totalRecords := 0
	comments := []*Comment{}

	for rows.Next() {

		var comment Comment

		err := rows.Scan(
			&totalRecords,
			&comment.ID,
			&comment.CreatedAt,
			&comment.Comment,
			&comment.Movie,
			&comment.CommenterIp,
		)

		if err != nil {
			return nil, 0,  err
		}

		comments = append(comments, &comment)
	}

	return comments, totalRecords,  nil
}


