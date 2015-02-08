package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"strings"
	"syscall"
	"text/template"

	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
	"github.com/nemith/gosx-notifier"
)

const (
	launchdLabel = "com.github.com.nemith.notifyserver"
)

var plistTmpl = template.Must(template.New("plist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
	<dict>
		<key>Label</key>
		<string>com.github.nemith.notifyserver</string>
		<key>ProgramArguments</key>
		<array>
			<string>{{.Path}}</string>
			<string>run</string>
			{{if .HTTPAddr}}
			<string>--http {{.HTTPAddr}}</string>
			{{end}}
		</array>
		<key>RunAtLoad</key>
		<true/>
		<key>KeepAlive</key>
		<true/>
		<key>SuccessfulExit</key>
		<true/>
	</dict>
</plist>
`))

type notification struct {
	Message  string `json:"message"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Group    string `json:"group"`
	Activate string `json:"activate"`
	Link     string `json:"link"`
	Sound    string `json:"sound"`
}

func serverAction(c *cli.Context) {
	r := mux.NewRouter()
	r.HandleFunc("/notify", notifyHandler).Methods("POST")

	fmt.Printf("Listening on: %s\n", c.String("http"))
	log.Fatal(http.ListenAndServe(c.String("http"), r))
}

func installAction(c *cli.Context) {
	binPath := c.String("bin")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		if binPath == "" {
			fmt.Fprintln(os.Stderr, "Cannot find binary for notifyserver.  Try 'go install github.com/nemith/notifyserver'")
		} else {
			fmt.Fprintf(os.Stderr, "Cannot find notifyserver binary at: '%s'\n", binPath)
		}
		os.Exit(1)
	}

	jobPath := launchdJobPath()
	f, err := os.OpenFile(jobPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	ctx := struct {
		Path     string
		HTTPAddr string
	}{
		Path:     c.String("bin"),
		HTTPAddr: c.String("http"),
	}
	err = plistTmpl.Execute(f, ctx)
	if err != nil {
		log.Fatal(err)
	}

	shortPath := strings.Replace(jobPath, userHomeDir(), "~", 1)

	fmt.Println("notifyserver launchd job has been installed and will start" +
		"automatically when you log in your system.\n")
	fmt.Printf("To start the notifyserver now:\n\tlaunchctl load  %[1]s\n"+
		"To stop the notifyserver:\n\tlaunchctl unload %[1]s\n", shortPath)
}

func uninstallAction(c *cli.Context) {
	jobPath := launchdJobPath()
	if err := os.Remove(jobPath); err != nil {
		if perr, ok := err.(*os.PathError); ok && perr.Err == syscall.ENOENT {
			fmt.Fprintln(os.Stderr, "notifyserver was not installed in launchd")
			return
		}
		panic(err)
	}
}

// notifyHandler will create osx notifcations for incoming HTTP posts
func notifyHandler(w http.ResponseWriter, r *http.Request) {
	var n notification
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&n); err != nil {
		http.Error(w, "Invalid json", 422)
		return
	}

	note := gosxnotifier.NewNotification(n.Message)
	note.Title = n.Title
	note.Subtitle = n.Subtitle
	note.Group = n.Group
	note.Activate = n.Activate
	note.Link = n.Link
	note.Sound = gosxnotifier.Sound(strings.ToTitle(n.Sound))

	if err := note.Push(); err != nil {
		log.Panic(err)
	}
}

func userHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

// launchdJobPath gets the path notifyserver's LaunchAgent for the current user.
func launchdJobPath() string {
	return fmt.Sprintf("%s/Library/LaunchAgents/%s.plist", userHomeDir(), launchdLabel)
}

// gopathBin will try to find notifyserver in the GOPATH/bin.
func gopathBin() string {
	gopaths := strings.Split(os.Getenv("GOPATH"), ":")
	for _, gopath := range gopaths {
		binPath := gopath + "/bin/notifyserver"
		if _, err := os.Stat(binPath); err == nil {
			return binPath
		}
	}
	return ""
}

func main() {
	app := cli.NewApp()
	app.Name = "notifyserver"
	app.Author = "Brandon Bennett"
	app.Email = "bennetb@gmail.com"
	app.Version = "0.0.1"
	app.Usage = "RESTful webserver to osx notifications"

	httpFlag := cli.StringFlag{
		Name:  "http, l",
		Value: "localhost:9999",
		Usage: "Use http and listen on the desired address",
	}

	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "Run the server in foreground",
			Action: serverAction,
			Flags: []cli.Flag{
				httpFlag,
			},
		},
		{
			Name:   "install",
			Usage:  "Install as launchd daemon",
			Action: installAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "bin, b",
					Value: gopathBin(),
					Usage: "Path to location of binary",
				},
				httpFlag,
			},
		},
		{
			Name:   "uninstall",
			Usage:  "Uninstall launchd daemon (if installed)",
			Action: uninstallAction,
		},
	}

	app.Run(os.Args)
}
