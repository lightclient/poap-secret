package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	badger "github.com/dgraph-io/badger/v3"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: poap-secret <links>\n")
		os.Exit(1)
	}

	db, err := badger.Open(badger.DefaultOptions("badger"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	seq, err := db.GetSequence([]byte{0x42}, 1)
	if err != nil {
		exit(err)
	}

	f, err := os.ReadFile(os.Args[1])
	if err != nil {
		exit(err)
	}

	links := strings.Split(string(f), "\n")

	// strip last link if empty
	if links[len(links)-1] == "" {
		links = links[:len(links)-1]
	}
	fmt.Printf("loaded %d links\n", len(links))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("form.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := t.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/input", func(w http.ResponseWriter, r *http.Request) {
		input := r.FormValue("input")
		if _, err := r.Cookie("week4"); err != http.ErrNoCookie {
			fmt.Fprintf(w, "already redeemed")
			return
		}
		if input == "contract" {
			next, err := seq.Next()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error getting next sequence: %v\n", err)
				fmt.Fprintf(w, "internal server error")
				return
			}

			// set cookie
			http.SetCookie(w, &http.Cookie{Name: "week4", Value: "true", HttpOnly: false})
			fmt.Fprintf(w, "%s", links[next])

			fmt.Printf("link %d / %d redeemed\n", next, len(links))

			return
		}
		fmt.Fprintf(w, "invalid secret: %s", input)
	})

	fmt.Println("listening on 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func exit(err error) {
	fmt.Fprintf(os.Stderr, "%v", err)
	os.Exit(1)
}
