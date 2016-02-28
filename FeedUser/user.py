import argparse, sys
import requests, json

def usage():
	print ("user <-i xx> [-s xx] [-u xx]")
	print ("Example usage: ")
	print ("		user -i 25 -s 3   ==>  subscribe user 25 to feed 3")
	print (" 	    user -i 25 -u 3   ==>  unsubscribe user 25 from feed 3")
	print ("		user -i 25 -p     ==>  print all feeds provided by server")
	print ("		user -i 25		  ==>  loop & get live feeds for 25")
	print ("		user -s 192.168.0.1 -i 25 ==>  instead of default localhost:8035, connects to server 192.168.0.1 and loop & get live feeds for 25")

def subscribe(ip, userid, feedid):
	print ("Subscribe {0} to feed {1}".format(userid, feedid))
	
	req_json = {"uid": userid, "fid": feedid }
	path = "{0}/subscribe".format(ip.rstrip('\\'));
	resp = requests.post(path, data=json.dumps(req_json),
                     headers={'Content-Type':'application/json'})
	if resp.status_code not in (200, 201):
		print("Status code", resp.status_code)
	print(resp.json())
	print("Done")

def unsubscribe(ip, userid, feedid):
	print ("Unubscribe {0} to feed {1}".format(userid, feedid))
	path = "{0}/unsubscribe".format(ip.rstrip('\\'));
	req_json = {"uid": userid, "fid": feedid }
	resp = requests.post(path, data=json.dumps(req_json),
                     headers={'Content-Type':'application/json'})
	if resp.status_code not in (200, 201):
		print("Status code", resp.status_code)
	print(resp.json())
	print("Done")

def print_supported_feeds(ip):
	path = "{0}/supported_feeds".format(ip.rstrip('\\'));
	resp = requests.get(path)
	if resp.status_code != 200:
		print("Status code", resp.status_code)
	print(resp.json())
	print("All supported feeds:")

def pollfeeds(ip, userid):
	print("Polling articles for user", userid)
	path = "{0}/articles/{1}".format(ip.rstrip('\\'), userid)
	print(path)
	resp = requests.get(path)
	if resp.status_code != 200:
		print("Status code", resp.status_code)
	for article in resp.json()["articles"]:
	    print('{}'.format(article))

def main():
	argparser = argparse.ArgumentParser(description="User for feed system")
	argparser.add_argument("-s", "--serveraddr", help="server address", default="http://localhost:8035")
	argparser.add_argument("-i", "--id", help='userid', required=True)
	argparser.add_argument("-p", "--printfeeds", help="print all supported feeds", action='store_true')
	group = argparser.add_mutually_exclusive_group()
	group.add_argument("-s", '--subscribe', help='feedid to subscribe to')
	group.add_argument('-u', '--unsubscribe', help='feedid to unsubscribe from')

	args = argparser.parse_args()
	print("Server ip addr ", args.serveraddr)

	userid = args.id
	if args.subscribe:
		subscribe(args.serveraddr, userid, args.subscribe)
	elif args.unsubscribe:
		unsubscribe(args.serveraddr, userid, args.unsubscribe)
	elif args.printfeeds:
		print_supported_feeds(args.serveraddr)
	else:
		pollfeeds(args.serveraddr, userid)
	return 0


if __name__ == "__main__":
	rc = main()
	sys.exit(rc)