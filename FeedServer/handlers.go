package main

import (
	"github.com/gorilla/mux"
	"encoding/json"
	"net/http"
	"log"
	"strconv"
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

	var result_channel chan string = make(chan string, 2) // small size just for receiving status
	cmd := user_feed_mutator_cmd{Cmdtype: 1, Uid: i.Uid, Fid: i.Fid, Result_channel: result_channel}
	user_feed_mutator_channel <- cmd

	status := <- result_channel
	if status == "OK" {
		u.Status = "OK"
	} else {
		u.Status = "UserSubscribe: Server Internal Error " + status
	}
	log.Printf("UserSubscribe: uid %s, fid %s, status %s", i.Uid, i.Fid, status)
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
	
	var result_channel chan string = make(chan string, 2)
	cmd := user_feed_mutator_cmd{Cmdtype: 2, Uid: i.Uid, Fid: i.Fid, Result_channel: result_channel}
	user_feed_mutator_channel <- cmd

	status := <- result_channel
	if status == "OK" {
		u.Status = "OK"
	} else {
		u.Status = "Server Internal Error " + status
	}
	log.Printf("UserUnsubscribe: uid %s, fid %s, status %s", i.Uid, i.Fid, status)
	return u
}


////////////////////////////////////////////////////////////////////////////////
type list_feeds_response struct {
	Fids 	[]string 	`json:"feeds"`
}

func GetSupportedFeeds(w http.ResponseWriter, r *http.Request) interface{} {
	u := new(list_feeds_response)
	
	var result_channel chan string = make(chan string, conf.FeedsResultChanDepth)
	cmd := article_feed_mutator_cmd{Cmdtype: 4, Aid: "-1", Fid: "-1", Result_channel: result_channel}
	article_feed_mutator_channel <- cmd

	tmp := <- result_channel
	cnt, err := strconv.Atoi(tmp) 
	if err != nil {
		log.Printf("Server Internal Error. Invalid feeds count %s", tmp)
	} else if cnt < 0 {
		log.Printf("Server Internal Error. Invalid feeds count %s", tmp)
	} else {
		u.Fids = make([]string, cnt)
		for i:=0; i < cnt; i++ {
			fid := <- result_channel
			u.Fids[i] = fid
		}
	}
	
	log.Printf("GetSupportedFeeds: got %d feeds", cnt)
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

	// #### step 1: get all feeds of this user
	fids :=	getFeedsForUser(uid)
	log.Printf("GetArticlesForUser: fids len %d", len(fids))
	
	// #### step 2: get all articles for each of the fid
	if len(fids) > 0 {
		//var articles []string
		for i:=0; i < len(fids); i++ {
			tmp := getArticlesForFeed(fids[i])//articles[:])
			if tmp != nil {
				u.ArticlesIds = append(u.ArticlesIds, tmp...)
			}
		}
		/*
		// Send back the list
		u.ArticlesIds = make([]string, len(articles))
		for i, article := range articles {
			u.ArticlesIds[i] = article
		}
		*/
	}

	log.Printf("GetArticlesForUser: uid %s, got %d articles", uid, len(u.ArticlesIds))
	return u
}

func getFeedsForUser(uid string) []string {
	var result_channel chan string = make(chan string, conf.FeedsResultChanDepth)
	cmd := user_feed_mutator_cmd{Cmdtype: 3, Uid: uid, Fid: "-1", Result_channel: result_channel}
	user_feed_mutator_channel <- cmd

	tmp := <- result_channel
	cnt, err := strconv.Atoi(tmp) 
	if err != nil {
		log.Printf("getFeedsForUser: Server Internal Error. Invalid feeds count %s", tmp)
		return nil
	} else if cnt < 0 {
		log.Printf("getFeedsForUser: Server Internal Error. %s", tmp)
		return nil
	} else {
		fids := make([]string, cnt)
		for i:=0; i < cnt; i++ {
			fid:= <- result_channel
			// we read as many times as possible to avoid blocking the mutator
			fids[i] = fid
		}
		log.Printf("getFeedsForUser: got %d feeds for user %s", len(fids), uid)
		return fids
	}
}

func getArticlesForFeed(fid string) []string {
	var result_channel chan string = make(chan string, conf.ArticlesResultChanDepth)
	cmd := article_feed_mutator_cmd{Cmdtype: 3, Aid: "-1", Fid: fid, Result_channel: result_channel}
	article_feed_mutator_channel <- cmd

	tmp := <- result_channel
	cnt, err := strconv.Atoi(tmp) 
	if err != nil {
		log.Printf("getArticlesForFeed: Server Internal Error. Invalid articles count %s", tmp)
		return nil
	} else if cnt < 0 {
		log.Printf("getArticlesForFeed: Server Internal Error. %s", tmp)
		return nil
	} else {
		articles := make([]string, cnt)
		for i := 0; i < cnt; i++ {
			aid := <- result_channel
			articles[i] = aid
		}
		log.Printf("getArticlesForFeed: got %d articles for feed %s", cnt, fid)
		return articles
	}
}

////////////////////////////////////////////////////////////////////////////////
type add_article_request struct {
	Suid	string		`json:"suid"`
	Fid 	string		`json:"fid"`
	Articles []string	`json:"articles"`
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
		dumpHttpRequest(r)
		return u
	}

	if !checkSUAuth(i.Suid) {
		log.Printf("AddArticleToFeed: auth failed %s", i.Suid)
		return nil
	}


	log.Printf("AddArticlesToFeed: suid %s, fid %s", i.Suid, i.Fid)
	for _, article := range i.Articles {
		var result_channel chan string = make(chan string, 1)
		cmd := article_feed_mutator_cmd{Cmdtype: 1, Aid: article, Fid: i.Fid, Result_channel: result_channel}
		article_feed_mutator_channel <- cmd
		status := <- result_channel
		if status != "OK" {
			u.Status = "Server Internal Error " + status
			return u
		}
	}

	u.Status = "OK"
	return u
}


////////////////////////////////////////////////////////////////////////////////
type get_feeds_of_user_response struct {
	Fids 	[]string 	`json:"feeds"`	
}

func GetFeedsOfUser(w http.ResponseWriter, r *http.Request) interface{} {
	vars := mux.Vars(r)
	if vars["suid"] == "" {
		log.Printf("GetArticlesForUser: invalid request, no suid. Ignore")
		return nil
	}
	if vars["uid"] == "" {
		log.Printf("GetArticlesForUser: invalid request, no uid. Ignore")
		return nil
	}

	suid := vars["suid"]
	if !checkSUAuth(suid) {
		log.Printf("GetArticlesForUser: auth failed %s", suid)
		return nil
	}

	uid := vars["uid"]
	u := new(get_feeds_of_user_response)

	u.Fids = getFeedsForUser(uid)
	
	log.Printf("GetFeedsOfUser: suid %s, uid %s, count of feeds %d", suid, uid, len(u.Fids))
	return u
}

////////////////////////////////////////////////////////////////////////////////
type get_users_of_feed_response struct {
	Uids 	[]string 	`json:"users"`	
}

func GetUsersOfFeed(w http.ResponseWriter, r *http.Request) interface{} {
	vars := mux.Vars(r)
	if vars["suid"] == "" {
		log.Printf("GetUsersOfFeed: invalid request, no suid. Ignore")
		return nil
	}
	if vars["fid"] == "" {
		log.Printf("GetUsersOfFeed: invalid request, no fid. Ignore")
		return nil
	}

	suid := vars["suid"]
	if !checkSUAuth(suid) {
		log.Printf("GetUsersOfFeed: auth failed %s", suid)
		return nil
	}

	fid := vars["fid"]
	u := new(get_users_of_feed_response)
	
	log.Printf("GetFeedsOfUser: suid %s, fid %s", suid, fid)
	return u
}