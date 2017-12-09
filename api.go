package main

import (
	"gorilla/mux"
	"html/template"
	"net/http"
)

type APIConfig struct {
	IPAddress string
	port      int
}

type Person struct {
	UserName string
}

type APIStatus struct {
	APIFiles   []APIFile
	TotalFiles int
	Percentage int
}

type APIFile struct {
	FileName string
	Uploaded bool
}

func api() {

	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/", files)
	r.HandleFunc("/temp", testtemplate)

	// Bind to a port and pass our router in
	http.ListenAndServe(":8000", r)
}

func files(w http.ResponseWriter, r *http.Request) {
	for _, folder := range config.Folders {
		folder.b2Files.b2GetCurrentFiles(*folder)
		listFiles := getFiles(folder.RootFolder)
		for _, file := range listFiles {
			var found bool
			for _, file2 := range folder.b2Files.Files {
				if (folder.B2Folder + file) == file2.FileName {
					found = true
				}
			}
			if !found {
				var newFile File
				newFile.RootPath = folder.RootFolder
				newFile.B2Path = folder.B2Folder
				newFile.FilePath = file[1:]
				newFile.BucketID = folder.BucketID
				w.Write([]byte(newFile.FilePath + "\n"))
			}
		}
	}
}

func testtemplate(w http.ResponseWriter, r *http.Request) {
	var fileStatus APIStatus
	var completedFiles, totalFiles int
	t := template.Must(template.ParseFiles("/Users/matthewarmstrong/scripts/backblaze-go/src/backblaze/static/main.html"))
	// files := fileCompleteQueue.Files
	for _, folder := range config.Folders {
		listFiles := getFiles(folder.RootFolder)
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
		fileStatus.TotalFiles = totalFiles
		fileStatus.Percentage = (completedFiles / totalFiles) //* 100
		t.Execute(w, fileStatus)
	}
}
