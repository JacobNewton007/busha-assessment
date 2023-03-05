package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type Result struct {
	Name   string `json:"name"`
	Height string `json:"height"`
	Gender string `json:"gender"`
}

type Metadata struct {
	Feets  float64 `json:"feets"`
	Inches float64 `json:"inches"`
	Count  int     `json:"count"`
}

type Character struct {
	Results []Result `json:"results"`
}

var character Character
var metadata Metadata

func (app *application) GetCharactersHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()

	sort_value := app.readString(qs, "sort", "")

	rdb := app.client
	charactersSortValue, _ := rdb.Get(Ctx, string(sort_value)).Result()

	filter_value := app.readString(qs, "gender", "")

	charactersFilterValue, _ := rdb.Get(Ctx, filter_value).Result()

	if charactersFilterValue != "" || charactersSortValue != "" {

		if charactersFilterValue != "" {
			character := Character{}
			json.Unmarshal([]byte(charactersFilterValue), &character)

			totalHeight := app.findTotalHeight(0, character)
			metadata = app.createCharacterMetaData(metadata, totalHeight, character)
			err := app.writeJSON(w, http.StatusOK, envelope{"character": character.Results, "metadata": metadata, "message": "fetch characters successfully", "status": "success"}, nil)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

		} else if charactersSortValue != "" {
			character := Character{}

			json.Unmarshal([]byte(charactersSortValue), &character)
			totalHeight := app.findTotalHeight(0, character)
			metadata = app.createCharacterMetaData(metadata, totalHeight, character)
			err := app.writeJSON(w, http.StatusOK, envelope{"character": character.Results, "metadata": metadata, "message": "fetch characters successfully", "status": "success"}, nil)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

		} else {
			characters, _ := rdb.Get(Ctx, sort_value).Result()
			character := Character{}

			json.Unmarshal([]byte(characters), &character)

			totalHeight := app.findTotalHeight(0, character)
			metadata = app.createCharacterMetaData(metadata, totalHeight, character)
			err := app.writeJSON(w, http.StatusOK, envelope{"character": character.Results, "metadata": metadata, "message": "fetch characters successfully", "status": "success"}, nil)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

		}

	} else {
		const url = "https://swapi.dev/api/people"
		body, err := app.fetchUrl(url)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		json.Unmarshal(body, &character)
		if filter_value == "" && sort_value != "" {

			sorted_character := app.sortBy(sort_value, character)

			data, err := json.Marshal(sorted_character)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			err = app.client.Set(Ctx, sort_value, data, time.Hour*2).Err()
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			totalHeight := app.findTotalHeight(0, character)
			metadata = app.createCharacterMetaData(metadata, totalHeight, sorted_character)
			err = app.writeJSON(w, http.StatusOK, envelope{"character": sorted_character.Results, "metadata": metadata, "message": "fetch characters successfully", "status": "success"}, nil)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

		} else if filter_value != "" && sort_value == "" {
			character, totalHeight := app.filterBy(filter_value, character)

			data, err := json.Marshal(character)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			err = app.client.Set(Ctx, filter_value, data, time.Hour*2).Err()
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			metadata = app.createCharacterMetaData(metadata, totalHeight, character)
			err = app.writeJSON(w, http.StatusOK, envelope{"character": character, "metadata": metadata, "message": "fetch characters successfully", "status": "success"}, nil)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

		} else {
			sorted_character := app.sortBy(sort_value, character)
			data, err := json.Marshal(sorted_character)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			err = app.client.Set(Ctx, "character", data, time.Hour*2).Err()
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			totalHeight := app.findTotalHeight(0, character)
			metadata = app.createCharacterMetaData(metadata, totalHeight, character)
			err = app.writeJSON(w, http.StatusOK, envelope{"character": sorted_character, "metadata": metadata, "message": "fetch characters successfully", "status": "success"}, nil)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

		}
	}
	// sort by params and filter by gender

}
