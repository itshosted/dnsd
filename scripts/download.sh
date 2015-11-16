#!/bin/bash
# Clone of https://github.com/jacobsalmela/pi-hole/blob/master/gravity.sh
#set -e
set -u
set -x

sources=('https://adaway.org/hosts.txt'
'http://adblock.gjtech.net/?format=unix-hosts'
#'http://adblock.mahakala.is/'
'http://hosts-file.net/ad_servers.txt'
'http://www.malwaredomainlist.com/hostslist/hosts.txt'
'http://pgl.yoyo.org/adservers/serverlist.php?'
'http://someonewhocares.org/hosts/hosts'
'http://winhelp2002.mvps.org/hosts.txt')

for ((i = 0; i < "${#sources[@]}"; i++))
do
	url=${sources[$i]}
	domain=$(echo "$url" | cut -d'/' -f3)
	file="../list.d/${domain}.txt"

	echo -n "Getting $domain list... "
	agent='Mozilla/5.0 (X11; Linux x86_64; rv:30.0) Gecko/20100101 Firefox/30.0'
	case "$domain" in
		"adblock.mahakala.is")
			cmd="curl -e http://forum.xda-developers.com/"
			;;

		"pgl.yoyo.org")
			cmd="curl -d mimetype=plaintext -d hostformat=hosts"
			;;

		# Default is a simple curl request
		*) cmd="curl"
	esac

	$cmd -A "$agent" $url > $file
done