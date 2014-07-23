package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/juju/errgo"
	"github.com/zenazn/goji/web"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func isErrSqlNoRows(err error) bool {
	return err == sql.ErrNoRows || err.(*errgo.Err).Underlying() == sql.ErrNoRows
}

func writeBody(w http.ResponseWriter, body []byte, status int, contentType string) error {
	w.Header().Set("Content-Type", contentType+"; charset=UTF8")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	_, err := w.Write(body)
	return errgo.Mask(err)
}

func renderJSON(w http.ResponseWriter, value interface{}, status int) error {
	body, err := json.Marshal(value)
	if err != nil {
		return errgo.Mask(err)
	}
	return writeBody(w, body, status, "application/json")
}

func renderString(w http.ResponseWriter, status int, msg string) error {
	return writeBody(w, []byte(msg), status, "text/plain")
}

func logError(err error) {
	s := fmt.Sprintf("Error:%s", err)
	if err, ok := err.(errgo.Locationer); ok {
		s += fmt.Sprintf(" %s", err.Location())
	}
	log.Println(s)
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	if err, ok := err.(HttpError); ok {
		http.Error(w, err.Error(), err.Status)
		return
	}

	if err, ok := err.(ValidationFailure); ok {
		renderJSON(w, err, http.StatusBadRequest)
		return
	}

	if isErrSqlNoRows(err) {
		http.NotFound(w, r)
		return
	}

	logError(err)

	http.Error(w, "Sorry, an error occurred", http.StatusInternalServerError)
}

func decodeJSON(r *http.Request, value interface{}) error {
	return errgo.Mask(json.NewDecoder(r.Body).Decode(value))
}

func scheme(r *http.Request) string {
	if r.TLS == nil {
		return "http"
	}
	return "https"
}

func baseURL(r *http.Request) string {
	return fmt.Sprintf("%s://%s", scheme(r), r.Host)
}

func getIntParam(c web.C, name string) int64 {
	value, _ := strconv.ParseInt(c.URLParams[name], 10, 0)
	return value
}

// Converts a Pg Array (returned as string) to an int slice
func pgArrToIntSlice(pgArr string) []int64 {
	var items []int64

	s := strings.TrimRight(strings.TrimLeft(pgArr, "{"), "}")

	for _, value := range strings.Split(s, ",") {
		if item, err := strconv.Atoi(value); err == nil {
			items = append(items, int64(item))
		}
	}
	return items
}

// Converts an int slice to a Pg Array string
func intSliceToPgArr(items []int64) string {
	var s []string
	for _, value := range items {
		s = append(s, strconv.FormatInt(value, 10))
	}
	return "{" + strings.Join(s, ",") + "}"
}

func getPage(r *http.Request) *Page {
	pageNum, err := strconv.ParseInt(r.FormValue("page"), 10, 64)
	if err != nil {
		pageNum = 1
	}
	return NewPage(pageNum)
}
