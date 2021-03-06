package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

type config struct {
	HttpPort int `json:"http_port"`
	CmdChanDepth int `json:"cmd_channel_depth"`
	FeedsResultChanDepth int `json:"feeds_result_channel_depth"`
	ArticlesResultChanDepth int `json:"articles_result_channel_depth"`
	Suid 	string `json:"su_id"`
	Datafile string `json:"datafile"`
}
var conf config

func commonWrapper(f func(http.ResponseWriter, *http.Request) interface{}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		u := f(w, r)
		b, _ := json.Marshal(u)
		w.Write(b)
	}
}

func main() {

	/* read configuration */
	file, err := os.Open("config.json")
	defer file.Close()

	if err != nil {
		/* configure not found */
		log.Printf("Unable to read config.json. Setting default parameters.")
		conf.HttpPort = 80
	} else {
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&conf)
		if err != nil {
			log.Printf("Error reading config.json: %s", err)
			return
		}
	}

	initialize()

	router := mux.NewRouter()
	// Each of these handler funcs would be called inside a go routine
	router.HandleFunc("/supported_feeds", commonWrapper(GetSupportedFeeds)).Methods("GET")
	router.HandleFunc("/subscribe", commonWrapper(UserSubscribe)).Methods("POST")
	router.HandleFunc("/unsubscribe", commonWrapper(UserUnsubscribe)).Methods("POST")
	router.HandleFunc("/articles/{uid:[0-9a-fA-F\\-]+}", commonWrapper(GetArticlesForUser)).Methods("GET")
	router.HandleFunc("/su/save_server_state", commonWrapper(SaveServerState)).Methods("POST")
	router.HandleFunc("/su/post_article", commonWrapper(AddArticleToFeed)).Methods("POST")
	router.HandleFunc("/su/get_feeds_of_user/{suid:[0-9a-fA-F\\-]+}/{uid:[0-9a-fA-F\\-]+}", commonWrapper(GetFeedsOfUser)).Methods("GET")
	//router.HandleFunc("/su/get_users_of_feed/{suid:[0-9a-fA-F\\-]+}/{fid:[0-9a-fA-F\\-]+}", commonWrapper(GetUsersOfFeed)).Methods("GET")
	
	http.Handle("/", router)

	log.Println(fmt.Sprintf("Listening at port %d ...", conf.HttpPort))
	http.ListenAndServe(fmt.Sprintf(":%d", conf.HttpPort), router)
	log.Println("Done! Exiting...")

	uninitialize()
}
