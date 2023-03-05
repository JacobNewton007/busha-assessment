package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/comments/:movie_name", app.MovieCommentsHandler)
	router.HandlerFunc(http.MethodPost, "/V1/comments/:movie_name", app.CreateCommentHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.GetMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/characters", app.GetCharactersHandler)

	return router
}
