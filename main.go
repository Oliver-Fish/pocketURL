package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type response struct {
	PocketURL string `json:"pocketURL"`
	Success   bool   `json:"success"`
	Msg       string `json:"msg"`
}

type pocketURL struct {
	RealURL string
	Created time.Time
}

/*
Todo:
	*Tests
	*Impliment SQLLite
*/

var pocketURLS = map[string]pocketURL{}

func main() {
	http.HandleFunc("/u/", getURL)
	http.HandleFunc("/add/", addURL)
	http.HandleFunc("/remove/", removeURL)
	fmt.Println("Starting Server on Port 8085")
	log.Fatal(http.ListenAndServe(":8085", nil)) //Define by startup args
}

func getURL(w http.ResponseWriter, req *http.Request) {
	surl := strings.Split(req.URL.String(), "/")[2] //Grab pURL code from URL
	if checkPocketURL(surl) {                       //Check if pURL code is valid
		http.Redirect(w, req, pocketURLS[surl].RealURL, 301) //Redirect to the RealURL of the pURL
		return
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 - URL Not Found"))

}
func addURL(w http.ResponseWriter, req *http.Request) {
	var resp response
	arg, f := req.URL.Query()["url"]
	if !f { //URL Query Empty
		resp = response{
			Success: false,
			Msg:     "URL Missing from request",
		}
		writeOutputJson(w, resp)
		return
	}
	url := arg[0]
	_, err := http.Get(url)
	if err != nil { //http.Get failed
		resp = response{
			Success: false,
			Msg:     "Unable to get requested url",
		}
		writeOutputJson(w, resp)
		return
	}
	pURL := checkURL(url)
	if pURL != "" { //URL exists return the existing pURL for it
		resp = response{
			PocketURL: pURL,
			Success:   true,
		}
		writeOutputJson(w, resp)
		return
	}
	var genURL string
	for genURL == "" && !checkPocketURL(genURL) { //Loop until we get a pocketurl not in use
		genURL = generatePocketURL(5)
	}
	pocketURLS[genURL] = pocketURL{ //Build our addURl Struct
		RealURL: url,
		Created: time.Now(),
	}

	resp = response{ //Build our Response Struct
		PocketURL: genURL,
		Success:   true,
	}
	writeOutputJson(w, resp) //Send Response to client
}

func removeURL(w http.ResponseWriter, req *http.Request) {
	var resp response
	pURL, p := req.URL.Query()["purl"]
	fURL, u := req.URL.Query()["url"]
	if p && u { //Both pURL and fURL have values this is not an accepted input
		resp = response{
			Msg:     "Expected purl or url got both",
			Success: false,
		}
		writeOutputJson(w, resp)
		return
	}
	if p { //pURL has been set
		url := pURL[0]
		pURLValue := checkPocketURL(url)
		if pURLValue { //URL exists
			delete(pocketURLS, url)
			resp = response{
				PocketURL: url,
				Msg:       "",
				Success:   true,
			}
			writeOutputJson(w, resp)
			return
		}
		resp = response{ //pURL not found
			Msg:     "Pocket URL not found",
			Success: false,
		}
		writeOutputJson(w, resp)
		return
	} else if u {
		url := fURL[0]
		pURLValue := checkURL(url)
		if pURLValue != "" { //URL exists
			delete(pocketURLS, pURLValue)
			resp = response{
				PocketURL: pURLValue,
				Success:   true,
			}
			writeOutputJson(w, resp)
			return
		}
		resp = response{ //URL not found
			Msg:     "URL not found",
			Success: false,
		}
		writeOutputJson(w, resp)
		return
	}
}

func checkPocketURL(purl string) bool {
	if _, ok := pocketURLS[purl]; ok {
		return true
	}
	return false
}

func checkURL(url string) string {
	for k, v := range pocketURLS {
		if v.RealURL == url {
			return k
		}
	}
	return "" //Return an empty string because not found
}

func writeOutputJson(w http.ResponseWriter, s interface{}) {
	dat, err := json.Marshal(&s)
	if err != nil {
		w.Write(dat)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

const randomCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func generatePocketURL(length int) string {
	rand.Seed(time.Now().UnixNano())
	r := make([]byte, length)
	for k := range r {
		r[k] = randomCharacters[rand.Intn(len(randomCharacters))]

	}
	return string(r)
}
