package main

import (
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"gopkg.in/mgo.v2"
)

type poll struct {
	ID      bson.ObjectId  `bson:"_id" json:"id"`
	Title   string         `json:"title" bson:"title"`
	Options []string       `json:"options"`
	Results map[string]int `json:"results,omitempty"`
}

func handlePolls(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handlePollsGet(w, r)
		return
	case "POST":
		handlePollsPost(w, r)
		return
	case "DELETE":
		handlePollsDelete(w, r)
		return
	case "OPTIONS":
		// Handles request to delete. Browser would request a permission to do so(pre-flight request)
		w.Header().Add("Access-Control-Allow-Methods", "DELETE")
		respond(w,r,http.StatusOK,nil)
		return
	}
	// handles unknown http method
	respondHTTPErr(w, r, http.StatusNotFound)
}

func handlePollsGet(w http.ResponseWriter, r *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("polls")
	var q *mgo.Query
	p := NewPath(r.URL.Path)
	if p.HasID() {
		// Search for specific id
		q = c.FindId(bson.ObjectIdHex(p.ID))
	} else {
		// Search all
		q = c.Find(nil)
	}
	var result []*poll
	if err := q.All(&result); err != nil {
		respondErr(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, &result)
}
func handlePollsPost(w http.ResponseWriter, r *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("polls")
	var p poll
	if err := decodeBody(r, &p); err != nil {
		respondErr(w, r, http.StatusBadRequest, "Cannot read searching options from request", err)
		return
	}
	p.ID = bson.NewObjectId()
	if err := c.Insert(p); err != nil {
		respondErr(w, r, http.StatusInternalServerError, "Failed storing searching options", err)
		return
	}
	w.Header().Set("Location", "polls/" + p.ID.Hex())
	respond(w, r, http.StatusCreated, nil)
}
func handlePollsDelete(w http.ResponseWriter, r *http.Request) {
	db := GetVar(r, "db").(*mgo.Database)
	c := db.C("polls")
	p := NewPath(r.URL.Path)
	if !p.HasID() {
		respondErr(w, r, http.StatusMethodNotAllowed, "Cannot delete all searching options")
		return
	}
	if err := c.Remove(bson.ObjectIdHex(p.ID)); err != nil {
		respondErr(w, r, http.StatusInternalServerError, "Failed deleteing searching otpions", err)
		return
	}
	respondErr(w, r, http.StatusOK, nil)
}
