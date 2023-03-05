package data

import ( 
	"database/sql"
	"errors"
)


// Define a custom ErrRecordNotFound error. We'll return this from our Get() method
// looking up a movie that doesn't exist in our database.

var (
	ErrRecordNotFound = errors.New("record not found")
)

// Create a models struct which wraps the commentsModel.
type Models struct {
	Comments CommentModels
}

func CommentFactory(db *sql.DB) Models {
	return Models{
		Comments: CommentModels{DB: db},
	}
}
