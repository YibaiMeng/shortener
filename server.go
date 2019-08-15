/*
    https://github.com/yibaimeng/shortener
    Meng Yibai <me@mengyibai.cn> 2019
    Released into the Public Domain
*/
package main 

import (
   "net/http"
   "net/url"
   "log"
   "fmt"
   "time"
   "math/rand"
   "errors"

   "github.com/boltdb/bolt"
   //"golang.org/x/net/idna" TODO: store the internationalized hostnames as punycode
)

var static_files = map[string]string{"" : "index.html", "index.html" : "index.html", "favicon.ico" : "favicon.ico", "robots.txt" : "robots.txt"}

type Storage = *bolt.DB
type Storage_config = string

func Init_storage(cfg Storage_config) (Storage, error) {
    db, err := bolt.Open(string(cfg), 0600, nil)
    if err != nil {
	    db.Close()
	    return db, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
	    _, _err := tx.CreateBucketIfNotExists([]byte("short2url"))
	    if _err != nil {
		    return fmt.Errorf("Create bucket: %s", err)
	    }
	    return nil
    })
    return db, err
}

func Store(src string, short string, db Storage) error {
    err := db.Update(func(tx *bolt.Tx) error {
	    b := tx.Bucket([]byte("short2url"))
	    if v := b.Get([]byte(short)); v != nil {
	        return errors.New("Store: shortend url already used.")
	    }
	    err := b.Put([]byte(short), []byte(src))
	    return err
    })
    return err
}

func Get(short string, db Storage) (string, error) {
    var src_url []byte
    err := db.View(func(tx *bolt.Tx) error {
	    b := tx.Bucket([]byte("short2url"))
	    src_url = b.Get([]byte(short))
	    if src_url == nil {
	        return errors.New("Get: No entry for shortend url.")
	    }
	    return nil
    })
    return string(src_url), err
}


func Shorten(src string, db Storage) string {
    // We use base58 to represent the shortend url. For each use, we generate a 63 bit random integer.
    // Every 6 bit generate a number.
    const base58 = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
    for {
        rand_int := rand.Int63()
        result := ""
        length := 3 // TODO: length need to change periodically. 
        for i := 0; i < length; i++ {
            result += string(base58[rand_int % 58])
            rand_int /= 58
        }
        if err := Store(src, result, db); err == nil {
            return result
        }
    }
}

func Expand(shortend string, db Storage) (string, error) {
    src_url, err := Get(shortend, db)
    return src_url, err
}

func gen_handler(db Storage) func(http.ResponseWriter, *http.Request) {
    return func (w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains") // Security considerations
        switch r.Method {
            case "POST": {
                log.Println("Received POST request.")
                log.Printf("r: %v.\n", r)
                auth := r.PostFormValue("auth")
                if auth != authkey {
                    log.Println("Auth error, not authorized.")
                    w.WriteHeader(http.StatusUnauthorized)
                    return
                }
                src_url, err := url.Parse(r.PostFormValue("url"))
                log.Printf("URL to shorten is %v.\n", src_url)
                if err != nil || (src_url.Host == "" && src_url.Path == "") {
                    w.WriteHeader(http.StatusBadRequest) // TODO: What headers to set?
                    log.Println("Request parameters invalid.")
                    return
	            } else {
	                if src_url.Scheme == "" {
                        src_url.Scheme = "https"
                    }
                    //parsed_src_url.Host, _ = idna.ToASCII(parsed_src_url.Host)
                    shortend := Shorten(src_url.String(), db)
                    log.Printf("URL shortend to %v.\n", shortend)
                    w.Write([]byte(domain+shortend+"\n"))
	            }
            }
            case "GET": {
                log.Println("Received GET request.")
                log.Printf("r: %v.\n", r)
                shortend := r.URL.Path[1:] // Everything after /
	            log.Printf("URL Path is %v.\n", shortend)
	            
	            if static_files[shortend] != "" {
	                http.ServeFile(w, r, static_files[shortend])
	                log.Println("Static file served")
	                return
	            }
	            
	            src_url, err := Expand(shortend, db)//Expand(code)
	            if err != nil {
	                log.Printf("URL Not Found. Error: %v\n", err)
	                w.WriteHeader(http.StatusNotFound)
	            } else {
	                log.Printf("URL expanded to %v.\n", src_url)
                    http.Redirect(w, r, src_url, 301)
                    log.Printf("301 redirect set.\n")
	            }
            }
        }
    }
}

func http_to_https(w http.ResponseWriter, r *http.Request) { 
    http.Redirect(w, r, "https://" + r.Host + r.RequestURI, http.StatusPermanentRedirect)
}

func main() {
    rand.Seed(time.Now().UnixNano())
    db, err := Init_storage(storage_cfg)
    if err != nil {
        log.Fatalf("Storage init unsuccessful: %v", err)
    }
	http.HandleFunc("/", gen_handler(db))
	// Must set up as a go routine!
    go http.ListenAndServeTLS(":443", cert_path, privkey_path, nil)
    log.Fatal(http.ListenAndServe(":80",  http.HandlerFunc(http_to_https)))
}
