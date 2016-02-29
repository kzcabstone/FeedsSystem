package main

import (
	"log"
	"strconv"
	"time"
	"encoding/json"
	"io/ioutil"
)

type user struct {
	Id 		string 		`json:"id"`
	Feeds	map[string]bool	`json:"feeds"`
}

type feed struct {
	Id		string				`json:"id"`
	//users	map[string]bool		`json:"users"`
	Articles map[string]int64	`json:"articles"`
}

type user_feed_mutator_cmd struct {
	Cmdtype int // 1: add user to feed   2: delete user from feed  3: get all feeds of a user 100: save all_users to json string
	Uid string
	Fid string
	Result_channel chan<- string // this channel(should be buffered) is passed in by user to us, we use it to send back result(feeds of the user in specific, one by one)
}

type article_feed_mutator_cmd struct {
	Cmdtype int // 1: add article to feed   3: get all articles of a feed  4: get all supported feeds
	Fid string
	Aid string
	Result_channel chan<- string
}

var all_users map[string]user
var all_feeds map[string]feed
var user_feed_mutator_channel chan user_feed_mutator_cmd
var user_feed_mutator_control_channel chan bool
var article_feed_mutator_channel chan article_feed_mutator_cmd
var article_feed_mutator_control_channel chan bool

// Initializes the maps and the channels
func initialize() {
	if !deserialize() {
		all_feeds = make(map[string]feed)
		all_users = make(map[string]user)
	}
	user_feed_mutator_channel = make(chan user_feed_mutator_cmd, conf.CmdChanDepth) // the channel only blocks if it's full
	user_feed_mutator_control_channel = make(chan bool) // control channel not buffered, so it's blocking
	article_feed_mutator_channel = make(chan article_feed_mutator_cmd, conf.CmdChanDepth)
	article_feed_mutator_control_channel = make(chan bool)

	// Start the mutators
	go user_feed_mutator()
	go article_feed_mutator()
}

func uninitialize() {
	t, e := serialize()
	if e != nil {
		log.Printf("uninitialize(): failed to dump state to file %s", conf.Datafile)
		return
	}
	log.Printf("uninitialize(): dumped state to file %s", conf.Datafile)
	log.Printf(string(t))
}

type serialized_server_state struct {
	Users string `json:"users"`
	Feeds string `json:"feeds"`
}

func serialize() ([]byte, error) {
	var strstate serialized_server_state

	var result_channel chan string = make(chan string, 2) // only 1 string expected to come out of this
	cmd := user_feed_mutator_cmd{Cmdtype: 100, Uid: "-1", Fid: "-1", Result_channel: result_channel}
	user_feed_mutator_channel <- cmd
	strstate.Users = <- result_channel

	var result_channel1 chan string = make(chan string, 2) // only 1 string expected to come out of this
	cmd1 := article_feed_mutator_cmd{Cmdtype: 100, Aid: "-1", Fid: "-1", Result_channel: result_channel1}
	article_feed_mutator_channel <- cmd1
	strstate.Feeds = <- result_channel1
	
	log.Printf(strstate.Users)
	log.Printf(strstate.Feeds)

	t, e := json.Marshal(strstate)

	err := ioutil.WriteFile(conf.Datafile, t, 0644)
	if err != nil {
		log.Printf("serialize(): writing to file %s failed", conf.Datafile)
	}

	return t, e
}

func deserialize() bool {
	dat, err := ioutil.ReadFile(conf.Datafile)
    if err != nil {
    	log.Printf("deserialize(): reading from file %s failed", conf.Datafile)
    	return false
    } else {
    	var strstate serialized_server_state
    	if e := json.Unmarshal(dat, &strstate); e != nil {
    		log.Printf("deserialize(): failed to unmarshal to strstate")
    		return false
    	} else {
    		e0 := json.Unmarshal([]byte(strstate.Users), &all_users)
    		e1 := json.Unmarshal([]byte(strstate.Feeds), &all_feeds)
    		if e0 != nil || e1 != nil {
    			log.Printf("deserialize(): failed to unmarshal to all_users/all_feeds")
    			return false
    		}
    	}
	    log.Printf(string(dat))
	    log.Printf("deserialize(): success. all_users.size: %d, all_feeds.size: %d", len(all_users), len(all_feeds))
	    return true
	}
}

// handles read/write to all_users and user.feeds
// TODO: use https://golang.org/src/sync/rwmutex.go to separate readers and writers of all_users map
// 			then we can have mutators to truely only serve mutatable requests
//          and let the reader to deal with all readonly request
func user_feed_mutator() {
	for {
		select {
			case cmd := <-user_feed_mutator_channel:
				processUserFeedCommand(cmd)
			case stop := <-user_feed_mutator_control_channel:
				if stop {
					log.Printf("Received stop signal. Exiting user_feed_mutator")
					return
				}
		}
	}
}

// handles read/write to all_feeds and feed.articles
func article_feed_mutator() {
	for {
		select {
			case cmd := <-article_feed_mutator_channel:
				processArticleFeedCommand(cmd)
			case stop := <-article_feed_mutator_control_channel:
				if stop {
					log.Printf("Received stop signal. Exiting user_feed_mutator")
					return
				}
		}
	}
}

func processUserFeedCommand(cmd user_feed_mutator_cmd) {
	_, userexists := all_users[cmd.Uid]
	if cmd.Cmdtype == 1 {
		// add user to feed		
		if !userexists {
			// add this user if not already there, this really belongs to register/login model which we don't have
			log.Printf("processUserFeedCommand(): user %s doesnot exist. create it", cmd.Uid)
			nuserptr := new(user)
			nuserptr.Id = cmd.Uid
			nuserptr.Feeds = make(map[string]bool)
			all_users[cmd.Uid] = *nuserptr
		} 
		_, feedsexists := all_users[cmd.Uid].Feeds[cmd.Fid]
		if feedsexists {
			// do nothing
			cmd.Result_channel <- "OK"
			return
		}
		// TODO: check if the feed is in all_feeds? Probably not necessary since no harm to add anyway
		all_users[cmd.Uid].Feeds[cmd.Fid] = true
		cmd.Result_channel <- "OK"
		log.Printf("processUserFeedCommand(): added user %s to feed %s", cmd.Uid, cmd.Fid)
	} else if cmd.Cmdtype == 2 {
		// delete user from feed
		if !userexists {
			// do nothing, send back error
			cmd.Result_channel <- "-1"
			log.Printf("processUserFeedCommand(): user %s doesnot exist. ignore", cmd.Uid)
			return
		}
		delete(all_users[cmd.Uid].Feeds, cmd.Fid)
		cmd.Result_channel <- "OK"
		log.Printf("processUserFeedCommand(): deleted user %s from feed %s", cmd.Uid, cmd.Fid)
	} else if cmd.Cmdtype == 3 {
		// get all feeds on the user
		if !userexists {
			cmd.Result_channel <- "-1"
			log.Printf("processUserFeedCommand(): user %s doesnot exist. nothing to send back", cmd.Uid)
			return
		}
		cntfeeds := 0
		cmd.Result_channel <- strconv.Itoa(len(all_users[cmd.Uid].Feeds))
		for fkey, _ := range all_users[cmd.Uid].Feeds {
			cmd.Result_channel <- fkey
			cntfeeds += 1
		}
		log.Printf("processUserFeedCommand(): sent back %d feeds for user %s", cntfeeds, cmd.Uid)
	} else if cmd.Cmdtype == 100 {
		jstr, err := json.Marshal(all_users)
		if err != nil {
			cmd.Result_channel <- "ERROR"
			log.Printf("processUserFeedCommand(): error marshaling all_users")
		} else {
			cmd.Result_channel <- string(jstr)
			log.Printf("processUserFeedCommand(): saved all users to json.")
		}
	} else {
		log.Printf("processUserFeedCommand(): unsupported cmd %d", cmd.Cmdtype)
		cmd.Result_channel <- "-1"
	}
}

func processArticleFeedCommand(cmd article_feed_mutator_cmd) {
	_, feedexists := all_feeds[cmd.Fid]
	if cmd.Cmdtype == 1 {
		// Add article to feed
		if !feedexists {
			// Create this feed
			log.Printf("processArticleFeedCommand(): feed %s doesnot exist. create it", cmd.Fid)
			nfeedptr := new(feed)
			nfeedptr.Id = cmd.Fid
			nfeedptr.Articles = make(map[string]int64)
			all_feeds[cmd.Fid] = *nfeedptr
		}
		all_feeds[cmd.Fid].Articles[cmd.Aid] = int64(time.Now().Unix())
		cmd.Result_channel <- "OK"
		log.Printf("processArticleFeedCommand(): added article %s to feed %s, timestamp %d", cmd.Aid, cmd.Fid, all_feeds[cmd.Fid].Articles[cmd.Aid])
	} else if cmd.Cmdtype == 3 {
		// Get all articles of a feed
		if !feedexists {
			cmd.Result_channel <- "-1"
			log.Printf("processArticleFeedCommand(): feed %s doesnot exist. ignore", cmd.Fid)
			return
		}
		cntarticles := 0
		cmd.Result_channel <- strconv.Itoa(len(all_feeds[cmd.Fid].Articles))
		for akey, _ := range all_feeds[cmd.Fid].Articles {
			cmd.Result_channel <- akey
			cntarticles += 1
		}
		log.Printf("processArticleFeedCommand(): sent back %d articles for feed %s", cntarticles, cmd.Fid)
	} else if cmd.Cmdtype == 4 {
		// Get all supported feeds
		cntfeeds := 0
		cmd.Result_channel <- strconv.Itoa(len(all_feeds))
		for fkey, _ := range all_feeds {
			cmd.Result_channel <- fkey
			cntfeeds += 1
		}
		log.Printf("processArticleFeedCommand(): get_all_supported_feeds sent back %d feeds", cntfeeds)
	} else if cmd.Cmdtype == 100 {
		jstr, err := json.Marshal(all_feeds)
		if err != nil {
			cmd.Result_channel <- "ERROR"
			log.Printf("processArticleFeedCommand(): error marshaling all_feeds")
		} else {
			cmd.Result_channel <- string(jstr)
			log.Printf("processArticleFeedCommand(): saved all feeds to json.")
		}
	} else {
		log.Printf("processArticleFeedCommand(): unsupported cmd %d", cmd.Cmdtype)
		cmd.Result_channel <- "-1"
	}
}
