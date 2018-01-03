package main

import (
	"gorilla/mux"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type APIConfig struct {
	StaticFiles string `yaml:"static_files"`
	BindIP      string `yaml:"bind_ip"`
	Port        string `yaml:"port"`
}

type Person struct {
	UserName string
}

type APIStatus struct {
	APIFiles   []APIFile
	TotalFiles float32
	Percentage float32
}

type APIFile struct {
	FileName string
	Uploaded bool
}

func api() {

	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/", files)
	r.HandleFunc("/status", status)
	r.HandleFunc("/log", readLog)

	// Bind to a port and pass our router in
	http.ListenAndServe((config.API.BindIP + ":" + config.API.Port), r)
}

func files(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Please see the documentation for using this api/webserver\n"))
}

func readLog(w http.ResponseWriter, r *http.Request) {
	file, _ := os.Open(filepath.Join(config.LogDir, "b2watcher.log"))
	defer file.Close()
	buf, _ := ioutil.ReadAll(file)
	w.Write(buf)

}

func status(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(filepath.Join(config.API.StaticFiles, "static/main.html")))
	// files := fileCompleteQueue.Files
	for _, folder := range config.Folders {
		var fileStatus APIStatus
		var completedFiles, totalFiles int
		w.Write([]byte("<h1>Folder " + folder.RootFolder + "</h1>\n"))
		listFiles := getFiles(folder.RootFolder)
		folder.b2Files.b2GetCurrentFiles(*folder)
		totalFiles = len(listFiles)
		completedFiles = 0
		for _, file := range listFiles {
			var fileStat APIFile
			var found bool
			for _, file2 := range folder.b2Files.Files {
				if (folder.B2Folder + file) == file2.FileName {
					found = true
				}
			}
			if !found {
				fileStat.Uploaded = false
			} else {
				completedFiles++
				fileStat.Uploaded = true
			}
			fileStat.FileName = file[1:]
			fileStatus.APIFiles = append(fileStatus.APIFiles, fileStat)
		}
		fileStatus.TotalFiles = float32(totalFiles)
		fileStatus.Percentage = (float32(completedFiles) / float32(totalFiles)) * 100
		t.Execute(w, fileStatus)
	}
}
