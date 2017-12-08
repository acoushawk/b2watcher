package main

import (
	"gorilla/mux"
	"net/http"
)

type APIConfig struct {
	IPAddress string
	port      int
}

func api() {

	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/", files)

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

func test(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Gorilla!\n"))
}
