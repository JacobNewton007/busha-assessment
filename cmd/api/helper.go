package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)


func (app *application) readMovieNameParams(r *http.Request) string  {
	params := httprouter.ParamsFromContext(r.Context())

	movie_name := params.ByName("movie_name")


	return movie_name
}

// Define an envelope type
type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {

	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body into the target destination.
	err := dec.Decode(dst)
	if err != nil {

		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var InvalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case errors.As(err, &InvalidUnmarshalError):
			panic(err)

		default:
			return err
		}

	}
	return nil
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {

	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *application) fetchUrl(url string) ([]byte, error) {
	client := http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	// var dsc  interface{}
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func (app *application) createCharacterMetaData(metadata Metadata, totalHeight int, character Character) Metadata {
	height_to_inches := (float64(totalHeight)/ 2.5) 
	height_to_feet := (float64(totalHeight)/ 30.48)
	metadata = Metadata{
		Feets: (math.Round(height_to_feet*100)/100),
		Inches: (math.Round(height_to_inches*100)/100),
		Count: len(character.Results),
	}

	return metadata
}

// sort either asc || desc 
func (app *application) sortBy(sort_value string, charactar Character) Character {
	if ( strings.HasPrefix(sort_value, "-")) {
		value := strings.TrimPrefix(sort_value, "-")
		switch value {
		case "name":
			sort.SliceStable(character.Results, func(i, j int) bool {
				return character.Results[i].Name > character.Results[j].Name
			})
		
		case "gender":
			sort.SliceStable(character.Results, func(i, j int) bool {
				return character.Results[i].Gender > character.Results[j].Gender
			})
	
		case "height":
			sort.SliceStable(character.Results, func(i, j int) bool {
				x, _ := strconv.ParseInt(charactar.Results[i].Height, 10, 64)
				y, _ := strconv.ParseInt(charactar.Results[j].Height, 10, 64)
				return int(x) >  int(y)
			})
		}
	} else {
		switch sort_value {
		case "name":
			sort.SliceStable(character.Results, func(i, j int) bool {
				return character.Results[i].Name < character.Results[j].Name
			})
		
		case "gender":
			sort.SliceStable(character.Results, func(i, j int) bool {
				return character.Results[i].Gender < character.Results[j].Gender
			})
	
		case "height":
			sort.SliceStable(character.Results, func(i, j int) bool {
				x, _ := strconv.ParseInt(charactar.Results[i].Height, 10, 64)
				y, _ := strconv.ParseInt(charactar.Results[j].Height, 10, 64)
				return int(x) <  int(y)
			})
		}
	}
	return charactar
}

func (app *application) filterBy(filter_value string, character Character ) (Character,  int) {
	totalHeight := 0
	var characters []Result

		switch filter_value {
			case "male":
				
				for i := range character.Results {
					if character.Results[i].Gender == "male" {
						characters = append(characters, character.Results[i])
						x, _ := strconv.ParseInt(character.Results[i].Height, 10, 64)
						totalHeight += int(x)
					}
				}

			case "female":
				for i := range character.Results {
					if character.Results[i].Gender == "female" {
						characters = append(characters, character.Results[i])
						x, _ := strconv.ParseInt(character.Results[i].Height, 10, 64)
						totalHeight += int(x)
					}
				}
		}
	
	character = Character{
		Results: characters,
	}

	return character, totalHeight
}


func (app *application) findTotalHeight(totalHeight int, character Character) int {
	for i := range character.Results {
		x, _ := strconv.ParseInt(character.Results[i].Height, 10, 64)
		totalHeight += int(x)
	}

	return totalHeight
}