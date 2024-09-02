package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var err error

// Song represents the song model in the existing music_song table
type Song struct {
	ID        uint   `gorm:"column:id"`
	Name      string `gorm:"column:name"`
	File      string `gorm:"column:file"`
	Author    string `gorm:"column:author"`
	Thumbnail string `gorm:"column:thumbnail"`
}

// TableName overrides the table name used by Gorm
func (Song) TableName() string {
	return "music_song"
}

func initDB() {
	// Initialize SQLite connection
	db, err = gorm.Open(sqlite.Open("../backend/db.sqlite3"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
}

func getSongID(r *http.Request) (int, error) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	return id, err
}

func getSongFromDB(id int) (Song, error) {
	var song Song
	err := db.First(&song, id).Error
	return song, err
}

func fetchFile(fileURL string) (*http.Response, error) {
	fullURL := "http://localhost:8000/media/" + fileURL
	resp, err := http.Get(fullURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file not found on the server")
	}
	return resp, nil
}

// Handles streaming of the file via HTTP range requests
func streamHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getSongID(r)
	if err != nil {
		log.Printf("Invalid song ID")
		http.Error(w, "Invalid song ID", http.StatusBadRequest)
		return
	}

	song, err := getSongFromDB(id)
	if err != nil {
		log.Printf("Song not found")
		http.Error(w, "Song not found", http.StatusNotFound)
		return
	}

	resp, err := fetchFile(song.File)
	if err != nil {
		log.Printf("Failed to fetch file")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	defer resp.Body.Close()

	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		log.Printf("serve file")
		http.ServeFile(w, r, song.File)
		return
	}

	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to parse file")
		return
	}

	reader := bytes.NewReader(buffer)
	http.ServeContent(w, r, song.Name, time.Now(), reader)
}

func main() {
	initDB()

	r := mux.NewRouter()
	r.HandleFunc("/songs/listen/{id}", streamHandler).Methods("GET")

	log.Println("Server is running on port 8005")
	log.Fatal(http.ListenAndServe(":8005", r))
}
