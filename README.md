#notifyserver
notifyserver is a service for OS X that listens for inbound HTTP json requests and will create OS notification for those requests. This is useful to send http notifications back across a tunneled ssh session or from a remote host with no crazy rpcs.

![Screenshot](https://raw.githubusercontent.com/nemith/notifyserver/_meta/ss1.png)

By default notifyserver listens on localhost:9999 and is intended to be used either locally (which is just silly) or over a tunneled ssh session.

## Installation and running

To install:

```
go get github.com/nemith/notifyserver
go install github.com/nemith/notifyserver
```

To run once in your terminal (assuming *$GOPATH/bin* is in your *$PATH*)

```
notifyserver run [--http <listen addr>]
```

To install as a launchd job for the current user you can run:

```
notifyserver install [--http <listen addr>]
```

To uninstall as a launchd run:

```
notifyserver uninstall
```

## Notify requests
To send notification to the server POST a json file with the following options to */notify*.  Only 'message' is required.

```json
{
 "message": "Yo you here?",
 "title": "IRC",
 "subtitle": "Message from bob",
 "sound": "Funk",
 "activate": "com.googlecode.iterm2",
 "group": "com.nemith.irc"
 }
```

### Example of sending a request

```
curl -H "Content-Type: application/json" -d @notify.json http://localhost:9999/notify
```


## Clients
You can find sample notifyserver clients in the clients directory

### Weechat 
There is a sample weechat plugin that can be used in conjunction with a ssh remoteforward to send weechat highlight and private message notifications to your desktop

Sample SSH config:
```
Host myvps
	Hostname myvps.vpshost.com
	RemoteForward 127.0.0.1:9999 127.0.0.1:9999
```


## TODO

 - [ ] Allow multipart images in post
 - [ ] Make weechat client more robust (away, current buffer, tmux awareness, etc)
