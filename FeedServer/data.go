package main

import (
	"log"
	"strconv"
	"time"
)

type user struct {
	id 		string 		`json:"id"`
	feeds	map[string]bool	`json:"feeds"`
}

type feed struct {
	id		string				`json:"id"`
	//users	map[string]bool		`json:"users"`
	articles map[string]int64	`json:"articles"`
}

type user_feed_mutator_cmd struct {
	Cmdtype int // 1: add user to feed   2: delete user from feed  3: get all feeds of a user
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

var all_users map[string]*user
var all_feeds map[string]*feed
var user_feed_mutator_channel chan user_feed_mutator_cmd
var user_feed_mutator_control_channel chan bool
var article_feed_mutator_channel chan article_feed_mutator_cmd
var article_feed_mutator_control_channel chan bool

// Initializes the maps and the channels
func initialize() {
	all_feeds = make(map[string]*feed)
	all_users = make(map[string]*user)
	user_feed_mutator_channel = make(chan user_feed_mutator_cmd, conf.CmdChanDepth) // the channel only blocks if it's full
	user_feed_mutator_control_channel = make(chan bool) // control channel not buffered, so it's blocking
	article_feed_mutator_channel = make(chan article_feed_mutator_cmd, conf.CmdChanDepth)
	article_feed_mutator_control_channel = make(chan bool)

	// Start the mutators
	go user_feed_mutator()
	go article_feed_mutator()
}

func uninitialize() {
	// free all memory
}

func serialize() {

}

func deserialize() {

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
			nuserptr.id = cmd.Uid
			nuserptr.feeds = make(map[string]bool)
			all_users[cmd.Uid] = nuserptr
		} 
		_, feedsexists := all_users[cmd.Uid].feeds[cmd.Fid]
		if feedsexists {
			// do nothing
			cmd.Result_channel <- "OK"
			return
		}
		// TODO: check if the feed is in all_feeds? Probably not necessary since no harm to add anyway
		all_users[cmd.Uid].feeds[cmd.Fid] = true
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
		delete(all_users[cmd.Uid].feeds, cmd.Fid)
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
		cmd.Result_channel <- strconv.Itoa(len(all_users[cmd.Uid].feeds))
		for fkey, _ := range all_users[cmd.Uid].feeds {
			cmd.Result_channel <- fkey
			cntfeeds += 1
		}
		log.Printf("processUserFeedCommand(): sent back %d feeds for user %s", cntfeeds, cmd.Uid)
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
			nfeedptr.id = cmd.Fid
			nfeedptr.articles = make(map[string]int64)
			all_feeds[cmd.Fid] = nfeedptr
		}
		all_feeds[cmd.Fid].articles[cmd.Aid] = int64(time.Now().Unix())
		cmd.Result_channel <- "OK"
		log.Printf("processArticleFeedCommand(): added article %s to feed %s, timestamp %d", cmd.Aid, cmd.Fid, all_feeds[cmd.Fid].articles[cmd.Aid])
	} else if cmd.Cmdtype == 3 {
		// Get all articles of a feed
		if !feedexists {
			cmd.Result_channel <- "-1"
			log.Printf("processArticleFeedCommand(): feed %s doesnot exist. ignore", cmd.Fid)
			return
		}
		cntarticles := 0
		cmd.Result_channel <- strconv.Itoa(len(all_feeds[cmd.Fid].articles))
		for akey, _ := range all_feeds[cmd.Fid].articles {
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
	} else {
		log.Printf("processArticleFeedCommand(): unsupported cmd %d", cmd.Cmdtype)
		cmd.Result_channel <- "-1"
	}
}
