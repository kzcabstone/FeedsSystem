import argparse, sys
import requests, json
import tokenize

def usage():
	print ("super_user <-i xx> [-s xx] [-u xx]")
	print ("Example usage: ")
	print ("		user -i 25 -l 3   ==>  list all feeds that user 3 subscribed")
	#print (" 	    user -i 25 -f 77   ==>  find all users that subscribed to 77")
	print ("		user -i 25 -p 77,a2,a7,a15	  ==>  post articles a2 a7 a15 to feed 77")

def listFeeds(ip, suid, userid):
	print ("List feeds for user {0}".format(userid))
	
	path = "{0}/su/get_feeds_of_user/{1}/{2}".format(ip.rstrip('\\'), suid, userid);
	print(path)
	resp = requests.get(path)
	if resp.status_code != 200:
		print("Status code", resp.status_code)
	print(resp.json())
	print("Done")

def findUsers(ip, suid, feedid):
	print ("Find users for feed {0}".format(feedid))
	
	path = "{0}/su/get_users_of_feed/{1}/{2}".format(ip.rstrip('\\'), suid, feedid);
	resp = requests.get(path)
	if resp.status_code != 200:
		print("Status code", resp.status_code)
	print(resp.json())
	print("Done")
	
def postArticle(ip, suid, param):
	firstcomma = param.index(',')
	feedid = param[0 : firstcomma]
	param = param[firstcomma+1 : ]
	articles = [x.strip() for x in param.split(',')]
	
	print ("postArticle {0} to feed {1}".format(articles, feedid))
	
	req_json = {"suid": suid, "fid": feedid, "articles": articles }
	path = "{0}/su/post_article".format(ip.rstrip('\\'));
	resp = requests.post(path, data=json.dumps(req_json),
                     headers={'Content-Type':'application/json'})
	if resp.status_code not in (200, 201):
		print("Status code", resp.status_code)
	print(resp.json())
	print("Done")


def main():
	argparser = argparse.ArgumentParser(description="Super User for feed system")
	argparser.add_argument("-s", "--serveraddr", help="server address", default="http://localhost:8035")
	argparser.add_argument("-i", "--id", help='superuserid', required=True)
	argparser.add_argument("-p", "--postarticle", help="<feedid,article1,article2,...> feedid is the feed to post these articles to")
	#group = argparser.add_mutually_exclusive_group()
	argparser.add_argument("-l", '--listfeeds', help='userid to list all feeds on')
	#group.add_argument('-f', '--findusers', help='feedid to list all subscribed users')

	args = argparser.parse_args()
	print("Server ip addr ", args.serveraddr)

	suid = args.id
	if args.listfeeds:
		listFeeds(args.serveraddr, suid, args.listfeeds)
	#elif args.findusers:
		#findUsers(args.serveraddr, suid, args.findusers)
	#	pass
	elif args.postarticle:
		postArticle(args.serveraddr, suid, args.postarticle)
	else:
		pass
	return 0


if __name__ == "__main__":
	rc = main()
	sys.exit(rc)