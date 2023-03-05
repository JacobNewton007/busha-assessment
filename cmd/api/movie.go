package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"
)

type Data struct {
	Title        string `json:"title"`
	OpeningCrawl string `json:"opening_crawl"`
	ReleaseDate  string `json:"release_date"`
	Date         time.Time `json:"-"`
	CommentCount int `json:"comment_count"`
}
type Movie struct {
	Results []Data `json:"results"`
}

func (app *application) GetMovieHandler(w http.ResponseWriter, r *http.Request) {
	var rdb = app.client
	var movie Movie
	movies, _ := rdb.Get(Ctx, "movies").Result()

	if movies != "" {
		data := Movie{}
    json.Unmarshal([]byte(movies), &data)
  
		err := app.writeJSON(w, http.StatusOK, envelope{"movies": data.Results}, nil)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

	} else {
		const url = "https://swapi.dev/api/films"

		body, err := app.fetchUrl(url)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		json.Unmarshal(body, &movie)

		for i := range movie.Results {
			date, err := time.Parse("2006-01-02", movie.Results[i].ReleaseDate)

			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			movie.Results[i].Date = date
			_, totalRecords, err := app.models.Comments.GetCommentForMovie(movie.Results[i].Title)
			if err != nil {
				app.notFoundResponse(w, r)
				return
			}

			movie.Results[i].CommentCount = totalRecords
		}

		sort.Slice(movie.Results, func(i, j int) bool { return movie.Results[i].Date.Before(movie.Results[j].Date) })
		data, err := json.Marshal(movie)

		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		err = rdb.Set(Ctx, "movies", data, time.Hour*2).Err()

		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		err = app.writeJSON(w, http.StatusOK, envelope{"movies": movie.Results, "message": "fetch movies successfully", "status": "success"}, nil)

		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

	}

}
