package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"gopkg.in/mgo.v2"

	"github.com/shurcooL/httpfs/html/vfstemplate"
)

// LoadTemplates 함수는 템플릿을 로딩합니다.
func LoadTemplates() (*template.Template, error) {
	t := template.New("")
	t, err := vfstemplate.ParseGlob(assets, t, "/template/*.html")
	return t, err
}

func webserver() {
	// 템플릿 로딩을 위해서 vfs(가상파일시스템)을 로딩합니다.
	vfsTemplate, err := LoadTemplates()
	if err != nil {
		log.Fatal(err)
	}
	TEMPLATES = vfsTemplate
	// assets 설정
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(assets)))
	// 웹주소 설정
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/search", handleSearch)
	http.HandleFunc("/add", handleAdd)
	// RestAPI
	http.HandleFunc("/api/add", handleAPIAdd)
	// 웹서버 실행
	http.ListenAndServe(*flagHTTPPort, nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	type recipe struct {
		Theme string `bson:"theme" json:"theme"`
	}
	rcp := recipe{
		Theme: "default.css",
	}
	q := r.URL.Query()
	userID := q.Get("userid")
	// 75mm studio 일때만 css 파일을 변경한다. 이 구조는 개발 초기에만 사용한다.
	if userID == "75mmstudio" {
		rcp.Theme = "75mmstudio.css"
	}
	err := TEMPLATES.ExecuteTemplate(w, "index", rcp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("add page"))
}

// handleSearch
func handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userID := q.Get("userid")
	year := q.Get("year")
	month := q.Get("month")
	day := q.Get("day")
	layer := q.Get("layer")
	sortKey := q.Get("sortkey")
	if userID == "" {
		http.Error(w, "URL에 userid를 입력해주세요", http.StatusBadRequest)
		return
	}

	log.Println(year, month, day, layer, sortKey)

	session, err := mgo.Dial(*flagDBIP)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	schedules, err := allSchedules(session, userID)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	err = json.NewEncoder(w).Encode(schedules)
	if err != nil {
		log.Println(err)
	}
}