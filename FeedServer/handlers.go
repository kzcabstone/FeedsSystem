package main

import (
	"github.com/gorilla/mux"
	"encoding/json"
	"net/http"
	"log"
)

////////////////////////////////////////////////////////////////////////////////
type subscription_request struct {
	Uid 	string 		`json:"uid"`
	Fid		string		`json:"fid"`
}

type subscription_response struct {
	Status  string     	`json:"status"`
}

func UserSubscribe(w http.ResponseWriter, r *http.Request) interface{} {
	jsondecoder := json.NewDecoder(r.Body)
	var i subscription_request
	u := new(subscription_response)
	if err := jsondecoder.Decode(&i); err != nil {
		u.Status = "Error"
		dumpHttpRequest(r);
		return u
	}
	u.Status = "OK"
	log.Printf("UserSubscribe: uid %s, fid %s", i.Uid, i.Fid)
	return u
}

func UserUnsubscribe(w http.ResponseWriter, r *http.Request) interface{} {
	jsondecoder := json.NewDecoder(r.Body)
	var i subscription_request
	u := new(subscription_response)
	if err := jsondecoder.Decode(&i); err != nil {
		u.Status = "Error"
		dumpHttpRequest(r);
		return u
	}
	u.Status = "OK"
	log.Printf("UserUnSubscribe: uid %s, fid %s", i.Uid, i.Fid)
	return u
}


////////////////////////////////////////////////////////////////////////////////
type list_feeds_response struct {
	Fids 	[]string 	`json:"feeds"`
}

func GetSupportedFeeds(w http.ResponseWriter, r *http.Request) interface{} {
	u := new(list_feeds_response)
	
	log.Printf("GetSupportedFeeds: nothing so far")
	return u
}


////////////////////////////////////////////////////////////////////////////////
type get_articles_request struct {
	Uid 	string 		`json:"uid"`
}

type get_articles_response struct {
	ArticlesIds	[]string	`json:"articles"`
}

func GetArticlesForUser(w http.ResponseWriter, r *http.Request) interface{} {
	vars := mux.Vars(r)
	if vars["uid"] == "" {
		log.Printf("GetArticlesForUser: invalid request, no userid. Ignore")
		return nil
	}

	uid := vars["uid"]
	u := new(get_articles_response)
	log.Printf("GetArticlesForUser: uid %s", uid)

	// Send back the list
	tmpArticles := [...]string{"article1", "article2", "article3"}
	u.ArticlesIds = make([]string, len(tmpArticles))
	for i, article := range tmpArticles {
		u.ArticlesIds[i] = article
	}

	return u
}


////////////////////////////////////////////////////////////////////////////////
type add_article_request struct {
	Suid	string		`json:"suid"`
	Fid 	string		`json:"fid"`
}

type add_article_response struct {
	Status  string     	`json:"status"`	
}

func AddArticleToFeed(w http.ResponseWriter, r *http.Request) interface{} {
	jsondecoder := json.NewDecoder(r.Body)
	var i add_article_request
	u := new(add_article_response)
	if err := jsondecoder.Decode(&i); err != nil {
		u.Status = "Error"
		dumpHttpRequest(r);
		return u
	}
	u.Status = "OK"
	log.Printf("AddArticleToFeed: suid %s, fid %s", i.Suid, i.Fid)
	return u
}


////////////////////////////////////////////////////////////////////////////////
type get_feeds_of_user_request struct {
	Suid	string		`json:"suid"`
	Uid 	string		`json:"uid"`
}

type get_feeds_of_user_response struct {
	Fids 	[]string 	`json:"feeds"`	
}

func GetFeedsOfUser(w http.ResponseWriter, r *http.Request) interface{} {
	jsondecoder := json.NewDecoder(r.Body)
	var i get_feeds_of_user_request
	u := new(get_feeds_of_user_response)
	if err := jsondecoder.Decode(&i); err != nil {
		dumpHttpRequest(r);
		return u
	}
	log.Printf("GetFeedsOfUser: suid %s, uid %s", i.Suid, i.Uid)
	return u
}