package main

import (
    "net/http"
    "io"
    "log"
    "os"
    "fmt"
    "path/filepath"
    "mime"
    "github.com/gorilla/mux"
    "bufio"
    "time"
    "github.com/lukin0110/push.kiwi/sanitize"
    html_template "html/template"
    "github.com/docker/go-units"
    "strconv"
    "path"
    "github.com/lukin0110/push.kiwi/utils"
    "errors"
    "strings"
    "net/url"
)

// 2 GigaByte
//const MAX_BYTES int = 2 * 1024 * 1024 * 1024
// 1 GigaByte
//const MAX_BYTES int64 = 1 * 1024 * 1024 * 1024
// 200 MegaByte
const MAX_BYTES int64 = 200 * 1024 * 1024
// 1 GigaByte: we need a least 1GB of free disk space
const MINIMAL_BYTES uint64 = 1 * 1024 * 1024 * 1024


func matcher(r *http.Request, rm *mux.RouteMatch) (match bool) {
    match = false

    var accept = r.Header.Get("Accept")
    //log.Printf("Accept header: %s", accept)

    if !strings.Contains(accept, "text/html") {
	return false
    }

    match = (r.Referer() == "")

    u, err := url.Parse(r.Referer())
    if err != nil {
	log.Fatal(err)
	return
    }

    match = match || (u.Path != r.URL.Path)
    return
}

func showPage(page string) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
	tpl_data, err := Asset(page)
	template, err := html_template.New(page).Parse(string(tpl_data[:]))
	    if err != nil {
	    http.Error(w, err.Error(), http.StatusInternalServerError)
	    return
	}

	data := struct {
	    Version html_template.HTML
	}{
	    html_template.HTML(fmt.Sprintf("<!-- %s -->", Full())),
	}

	if err := template.Execute(w, data); err != nil {
	    http.Error(w, err.Error(), http.StatusInternalServerError)
	    return
	}
    }
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "404 Page not found - Ceci n'est pas une page", 404)
}

func showHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)

    token := vars["token"]
    filename := vars["filename"]

    log.Printf("View file: %s/%s\n", token, filename)
    var fullPath = filepath.Join("/storage", token, filename)
    file, err := os.Open(fullPath)
    defer file.Close()

    if err != nil {
	log.Printf("%s", err.Error())
	http.Error(w, "File not found", 404)
	return
    }

    fileStats, err := file.Stat()

    //reader, contentType, contentLength, err := storage.Get(token, filename)
    reader := bufio.NewReader(file)

    // Detect the content type of a file
    var contentType string
    testBytes, err := reader.Peek(64) //read a few bytes without consuming
    if err != nil {
	contentType = mime.TypeByExtension(fullPath)
	log.Printf("%s", err.Error())
	//http.Error(w, "Could not determine Content-Type", http.StatusInternalServerError)
	//return
    } else {
    	contentType = http.DetectContentType(testBytes)
    }

    var content html_template.HTML
    var humanSize = units.HumanSize(float64(fileStats.Size()))

    tpl_data, err := Asset("static/download.html")
    template, err := html_template.New("download").Parse(string(tpl_data[:]))

    if err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
    }

    data := struct {
	ContentType	string
	Content		html_template.HTML
	Filename	string
    	Url		string
	HumanSize	string
	Version		html_template.HTML
    }{
	contentType,
	content,
	filename,
	r.URL.String(),
	humanSize,
	html_template.HTML(fmt.Sprintf("<!-- %s -->", Full())),
    }

    if err := template.Execute(w, data); err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
    }
}

func previewEmailHandler(w http.ResponseWriter, r *http.Request) {
    tpl_data, _ := Asset("static/email.html")
    template, _ := html_template.New("email").Parse(string(tpl_data[:]))

    data := struct {
	Url         string
	HumanSize   string
	DeletedOn   string
    }{
	"https://push.kiwi/ubbTc74ul/foobar.jpg",
	"44.91 kB",
	"Wednesday 16 November, 2016 at 14:16 (UTC)",
    }

    if err := template.Execute(w, data); err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func putHandler(w http.ResponseWriter, r *http.Request) {
    // First, check the disk usage to avoid a full disk :)
    du := utils.NewDiskUsage("/")
    if du.Free() < MINIMAL_BYTES {
	log.Printf("No space left on disk. Available: %s", units.HumanSize(float64(du.Free())))
	http.Error(w, errors.New("No space left on disk :(").Error(), http.StatusInternalServerError)
	return
    } else {
	log.Printf("Available disk space: %s", units.HumanSize(float64(du.Free())))
    }

    vars := mux.Vars(r)
    var filename string = sanitize.Path(filepath.Base(vars["filename"]))
    var reader io.Reader = r.Body
    var contentType string = r.Header.Get("Content-Type")
    var contentLength string = r.Header.Get("Content-Length")

    // Print all headers for bedugging
    for k, v := range r.Header {
	log.Println("key:", k, "value:", v)
    }

    if contentLength != "" {
	log.Printf("Content length %s\n", contentLength)
	byteSize, err1 := strconv.ParseInt(contentLength, 10, 64)
	if err1 == nil && byteSize > MAX_BYTES {
	    http.Error(w, fmt.Errorf("File size is too large %s", units.HumanSize(float64(byteSize))).Error(), http.StatusInternalServerError)
	    return
	}
    }

    if contentType == "" {
	contentType = mime.TypeByExtension(filepath.Ext(vars["filename"]))
    }

    var exists bool = true
    var token string
    var fullPath string

    for exists {
	token = utils.RandStringBytesMaskImprSrc(10)
	fullPath = path.Join("/storage", token, filename)
	_, err := os.Stat(fullPath)
	log.Printf("File check: %s", err)
	exists = os.IsExist(err)
    }

    log.Printf("Uploading: %s %s %s ...", token, filename, contentType)
    size, err1 := writeFile(reader, fullPath, MAX_BYTES)
    if err1 != nil {
	// Cleanup the file (and it's token directory) if an error happened
	defer func() {
	    os.Remove(fullPath)
	    defer os.Remove(path.Join("/storage", token))
	}()
	log.Printf("%s", err1.Error())
	http.Error(w, err1.Error(), http.StatusInternalServerError)
	return
    }

    var url = fmt.Sprintf("%s/%s/%s", config.RootUrl, token, filename)
    var email = r.Header.Get("x-email")

    if email != "" {
	log.Printf("Emailing '%s' to: %s", url, email)
	go func() {
	    // Delete the file in 14 hours
	    t := time.Now().Add(time.Hour * 24)
	    err := mail(email, filename, url, float64(size), t)
	    if err != nil {
		log.Println(err)
	    }
	}()
    }

    w.Header().Set("Content-Type", "text/plain")
    fmt.Fprintf(w, "%s\n", url)
}


// http://stackoverflow.com/questions/1821811/how-to-read-write-from-to-file
func writeFile(reader io.Reader, fullPath string, maxSize int64) (size int, err error) {
    // make a read buffer
    r := bufio.NewReader(reader)

    // open output file
    size = 0
    log.Printf("Storing: %s ...", fullPath)
    os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)	// Create base directory if it doesn't exist

    fo, err := os.Create(fullPath)
    if err != nil {
        return
    }

    // close fo on exit and check for its returned error
    defer func() {
        if err := fo.Close(); err != nil {
            return
        }
    }()
    // make a write buffer
    w := bufio.NewWriter(fo)

    // make a buffer to keep chunks that are read
    buf := make([]byte, 1024)
    for {
        // read a chunk
        n, err := r.Read(buf)
        if err != nil && err != io.EOF {
            return -1, err
        }

	size += n
	if int64(size) > maxSize {
	    err := fmt.Errorf("Max file size exceeded %s", units.HumanSize(float64(maxSize)))
	    return -1, err
	}

        if n == 0 {
            break
        }

        // write a chunk
        if _, err := w.Write(buf[:n]); err != nil {
            return -1, err
        }
    }

    if err = w.Flush(); err != nil {
        return
    }

    return
}

func getHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)

    token := vars["token"]
    filename := vars["filename"]
    log.Printf("Download file: %s/%s\n", token, filename)

    var fullPath = filepath.Join("/storage", token, filename)
    file, err := os.Open(fullPath)
    defer file.Close()

    if err != nil {
	http.Error(w, "File not found", 404)
	return
    }

    fileStats, err := file.Stat()
    if err != nil {
	log.Printf("%s", err.Error())
	http.Error(w, "Could not fetch stats of file", http.StatusInternalServerError)
	return
    }

    //reader, contentType, contentLength, err := storage.Get(token, filename)
    reader := bufio.NewReader(file)

    // Detect the content type of a file
    var contentType string
    testBytes, err := reader.Peek(64) //read a few bytes without consuming
    if err != nil {
	contentType = mime.TypeByExtension(fullPath)
	//http.Error(w, fmt.Sprintf("Could not determine content-type, %s", err), http.StatusInternalServerError)
	//return
    } else {
	contentType = http.DetectContentType(testBytes)
    }

    w.Header().Set("Content-Type", contentType)
    w.Header().Set("Content-Length", strconv.FormatInt(fileStats.Size(), 10))
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
    w.Header().Set("Connection", "close")

    if _, err = io.Copy(w, reader); err != nil {
	log.Printf("%s", err.Error())
	http.Error(w, "Error occurred copying to output stream", http.StatusInternalServerError)
	return
    }
}
