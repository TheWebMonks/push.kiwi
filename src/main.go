package main

import (
    "fmt"
    "net/http"
    "os"
    "github.com/gorilla/mux"
    "flag"
    "log"
    "os/signal"
    "syscall"
    "runtime"
    "mime"
    "github.com/lukin0110/push.kiwi/utils"
    "github.com/robfig/cron"
)

var BUILD_DATE string

type Config struct {
    RootUrl		string
    Temp		string
    MailgunDomain	string
    MailgunKey		string
    MailgunPublicKey	string
}

func (c Config) String() string {
    return fmt.Sprintf("{RootUrl: %s, Temp: %s, MailgunDomain: %s, MailgunKey: %s, MailgunPublicKey: %s}",
	c.RootUrl,
	c.Temp,
	c.MailgunDomain,
	c.MailgunKey,
	c.MailgunPublicKey)
}

var config = Config{}

func init() {
    config.Temp = os.TempDir()
}

func redirect(w http.ResponseWriter, req *http.Request) {
    http.Redirect(w, req, "https://" + req.Host + req.URL.String(), http.StatusMovedPermanently)
}

//Generate templates data
//go:generate go-bindata static/...
func main() {
    clean := flag.Bool("clean", false, "Execute storage cleanup")
    version := flag.Bool("version", false, "Show version")
    tls := flag.Bool("tls", false, "Use TLS")
    tlsCrt := flag.String("tls-cert", "", "SSL Certificate file")
    tlsKey := flag.String("tls-key", "", "SSL Key file")
    rootUrl := flag.String("root-url", "https://push.kiwi", "Root url to use in e-mails")
    mailgunDomain := flag.String("mailgun-domain", "push.kiwi", "Mailgun domain to send emails from")
    mailgunKey := flag.String("mailgun-key", "", "")
    mailgunPublicKey := flag.String("mailgun-public-key", "", "")
    flag.Parse()

    config.RootUrl = *rootUrl
    config.MailgunDomain = *mailgunDomain
    config.MailgunKey = *mailgunKey
    config.MailgunPublicKey = *mailgunPublicKey

    if *clean {
	fmt.Println("Cleaning /storage")
	utils.CleanStorage("/storage")
	os.Exit(0)
    }

    if *version {
	fmt.Printf("Version %s, build date: %s\n", Full(), BUILD_DATE)
	os.Exit(0)
    }

    // Limiting the maximum threads
    numberCPU := runtime.NumCPU()
    runtime.GOMAXPROCS(numberCPU)
    log.Printf("Config: %s", config)

    r := mux.NewRouter()
    r.HandleFunc("/{filename}", putHandler).Methods("PUT")
    r.HandleFunc("/{token}/{filename}", showHandler).MatcherFunc(matcher).Methods("GET")
    r.HandleFunc("/{token}/{filename}", getHandler).Methods("GET")
    r.HandleFunc("/email", previewEmailHandler).Methods("GET")
    r.HandleFunc("/kiwipedia.html", showPage("static/kiwipedia.html")).Methods("GET")
    r.HandleFunc("/", showPage("static/index.html")).Methods("GET")
    r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

    http.Handle("/", r)
    s := &http.Server{
	Addr:    fmt.Sprintf(":%s", "443"),
    }

    mime.AddExtensionType(".md", "text/x-markdown")
    mime.AddExtensionType(".txt", "text/plain")

    if *tls {
	//log.Fatal(http.ListenAndServe(":8080", nil))
	go func() {
	    log.Printf("Push.Kiwi server started. \n\tlistening on port: %v\n\tusing temp folder: %s\n", "443", config.Temp)
	    if !utils.Exists(*tlsCrt) {
		log.Fatal(fmt.Sprintf("--tls-crt '%s' doesn't exist", *tlsCrt))
		os.Exit(-1)
	    }

	    if !utils.Exists(*tlsKey) {
		log.Fatal(fmt.Sprintf("--tls-key '%s' doesn't exist", *tlsKey))
		os.Exit(-1)
	    }

	    log.Fatal(s.ListenAndServeTLS(*tlsCrt, *tlsKey))
	}()

	go func() {
	    log.Printf("Push.Kiwi server started. \n\tlistening on port: %v\n", "80")
	    http.ListenAndServe(":80", http.HandlerFunc(redirect))
	}()
    } else {
	go func() {
	    port := "8080"
	    log.Printf("Push.Kiwi server started. \n\tlistening on port: %v\n\tusing temp folder: %s\n", port, config.Temp)
	    log.Fatal(http.ListenAndServe(":" + port, nil))
	}()
    }

    //https://godoc.org/github.com/robfig/cron
    cronny := cron.New()
    //cronny.AddFunc("0 1 * * * *", func() { fmt.Println("Every hour on the half hour") })
    //cronny.AddFunc("@every 1m", func() { fmt.Println("Every minute1"); log.Println("Every minute2") })
    cronny.AddFunc("@daily", func() { utils.CleanStorage("/storage") })
    cronny.Start()

    term := make(chan os.Signal, 1)
    signal.Notify(term, os.Interrupt)
    signal.Notify(term, syscall.SIGTERM)

    <-term
    log.Print("Server stopped.")
}
