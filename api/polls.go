package main

import (
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"errors"
)

type poll struct {
	ID bson.ObjectId `bson: "_id" json: "id"`
	Title string `json: "title"`
	Options []string `json: "options"`
	Results map[string]int `json: "results, omitempty"`
}

func handlePolls(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handlePollsGet(w,r)
		return
	case "POST":
		handlePollsPost(w,r)
		return
	case "DELETE":
		handlePollsDelete(w,r)
		return
	}
	// handles unknown http method
	respondHTTPErr(w, r, http.StatusNotFound)
}

func handlePollsGet(w http.ResponseWriter, r *http.Request) {
	respondErr(w, r, http.StatusInternalServerError, errors.New("Not implemented"))
}
func handlePollsPost(w http.ResponseWriter, r *http.Request) {
	respondErr(w, r, http.StatusInternalServerError, errors.New("Not implemented"))
}
func handlePollsDelete(w http.ResponseWriter, r *http.Request) {
	respondErr(w, r, http.StatusInternalServerError, errors.New("Not implemented"))
}
