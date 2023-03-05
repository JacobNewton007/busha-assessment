package api

import (
	"fmt"
	"net/http"
	"time"

	custom_validator "github.com/JacobNewton007/busha-test/internals/custom_validator"
	"github.com/JacobNewton007/busha-test/internals/data"
	"github.com/go-playground/validator/v10"
)

func getClientIpAddr(req *http.Request) string {
	clientIp := req.Header.Get("X-FORWARDED-FOR")
	if clientIp != "" {
		return clientIp
	}
	return req.RemoteAddr
}

func (app *application) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Comment     string `json:"comment" validate:"required,min=4,max=500"`
		Movie       string `json:"movie_name" validate:"required"`
		CommenterIp string `json:"commenter_ip" validate:"required"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	input.Movie = app.readMovieNameParams(r)
	input.CommenterIp = getClientIpAddr(r)

	comment := &data.Comment{
		Comment:     input.Comment,
		Movie:       input.Movie,
		CommenterIp: input.CommenterIp,
	}

	validate := validator.New()

	trans := custom_validator.Validator()

	err = validate.Struct(comment)
	if err != nil {
		errs := custom_validator.TranslateError(err, trans)

		app.logger.Error(errs)
	}

	err = app.models.Comments.Insert(comment)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/comments/%s", comment.Movie))
	// set movies key to expire when new comment is added to a movie
	app.client.Expire(Ctx, "movies", time.Second * 1)

	err = app.writeJSON(w, http.StatusCreated, envelope{"comment": comment, "message": "comment created", "status": "success"}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) MovieCommentsHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Movie string `json:"movie_name"`
	}

	input.Movie = app.readMovieNameParams(r)

	comments, totalRecords, err := app.models.Comments.GetCommentForMovie(input.Movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comments": comments, "totalRecords": totalRecords, "message": "fetch comment successfully", "status": "success"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
