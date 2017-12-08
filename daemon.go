package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"math"
	"net/url"
	"os"
	"time"
)

var getSHAChan = make(chan File, 5)
var processFileChan = make(chan File, 5)
var completedFileChan = make(chan FilePart)
var uploadFilePart = make(chan FilePart)
var exitChan = make(chan bool)
var fileCompleteQueue FileQueue

func daemon() {
	var monitor bool
	f, err := os.OpenFile((config.LogDir + "/b2watcher.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	log.SetOutput(f)
	for i := 1; i <= config.ConnConnections; i++ {
		go sendFilePart()
	}
	go getSHA()
	go sendFile()
	// go api()
	for _, folder := range config.Folders {
		if folder.Monitor == true {
			go folderMonitor(folder)
			monitor = true
		}
	}
	if !monitor {
		go queueMonitor()
	}
	for {
		select {
		case filePart := <-completedFileChan:
			fileCompleteQueue.updateFile(filePart)
			for _, file := range fileCompleteQueue.Files {
				complete := true
				for _, part := range file.Parts {
					if part.Complete == false {
						complete = false
					}
				}
				if (complete == true) && len(file.Parts) == 1 {
					log.Println("Finished sending file ", file.FilePath)
					fileCompleteQueue.removeFile(file)
				} else if complete == true {
					log.Println("Finished sending file ", file.FilePath)
					file.b2FinishLargeFile()
					fileCompleteQueue.removeFile(file)
				}
			}
		case <-exitChan:
			log.Println("Finished processing files, no folders set to monitor. Closing")
			os.Exit(0)
		}
	}
}

func getSHA() {
	for {
		select {
		case file := <-getSHAChan:
			log.Println("Starting File - ", file.FilePath)
			f, err := os.Open(file.RootPath + "/" + file.FilePath)
			if err != nil {
				log.Fatal(err)
			}
			info, _ := f.Stat()
			file.FileSize = info.Size()
			if file.FileSize > instance.RecPartSize {
				var bytesSent int64
				chunkSize := instance.RecPartSize
				totalPartsNum := uint64(math.Ceil(float64(file.FileSize) / float64(instance.RecPartSize)))
				for i := int64(0); i < int64(totalPartsNum); i++ {
					log.Println("Working on part", i, " of file ", file.FilePath)
					var filePart FilePart
					var buffer []byte
					if (file.FileSize - bytesSent) < chunkSize {
						chunkSize = file.FileSize - bytesSent
					}
					f.Seek((bytesSent), 0)
					buffer = make([]byte, chunkSize)
					h := sha1.New()
					f.Read(buffer)
					if _, err := io.Copy(h, bytes.NewReader(buffer)); err != nil {
						// log.Fatal(err)
						fmt.Println(err)
					}
					fileSHA := fmt.Sprintf("%x", h.Sum(nil))

					filePart.ChunkSize = chunkSize
					filePart.Number = i + 1
					filePart.SHA = fileSHA
					file.Parts = append(file.Parts, filePart)
					bytesSent = bytesSent + chunkSize
				}
			} else {
				var filePart FilePart
				h := sha1.New()
				if _, err := io.Copy(h, f); err != nil {
					log.Fatal(err)
				}
				fileSHA := fmt.Sprintf("%x", h.Sum(nil))
				filePart.SHA = fileSHA
				filePart.ChunkSize = file.FileSize
				filePart.Number = 1
				file.Parts = append(file.Parts, filePart)
			}
			h := sha1.New()
			if _, err := io.Copy(h, f); err != nil {
				log.Fatal(err)
			}
			fileSHA := fmt.Sprintf("%x", h.Sum(nil))
			file.SHA = fileSHA
			log.Println("Starting upload of ", file.FilePath)
			processFileChan <- file
			f.Close()
		}
	}
}

func sendFile() {
	for {
		select {
		case file := <-processFileChan:
			if len(file.Parts) == 1 {
				file.b2UploadURL()
				fileCompleteQueue.addFile(file)
				for _, part := range file.Parts {
					part.ParentFileID = file.FileID
					part.Path = file.RootPath + "/" + file.FilePath
					part.URL = file.UploadURL.UploadURL
					part.AuthToken = file.UploadURL.AuthorizationToken
					b2FileName, _ := url.Parse(file.B2Path + "/" + file.FilePath)
					part.FileName = b2FileName.String()
					uploadFilePart <- part
				}
			} else {
				file.b2StartLargeFile()
				// for i := 1; i <= len(file.Parts); i++ {
				// 	file.b2UploadPartURL()
				// }
				fileCompleteQueue.addFile(file)
				for _, part := range file.Parts {
					part.ParentFileID = file.FileID
					part.Path = file.RootPath + "/" + file.FilePath
					uploadFilePart <- part
				}
			}
		}
	}
}

func sendFilePart() {
	for {
		select {
		case filePart := <-uploadFilePart:
			var result int
			if (filePart.Number == 1) && (filePart.ChunkSize < instance.RecPartSize) {
				result = filePart.b2UploadFile()
			} else {
				filePart.b2UploadPartURL()
				result = filePart.b2UploadPart()
			}
			if result == 200 {
				log.Println("Finished Part ", filePart.Number, " of file ", filePart.Path)
				filePart.Complete = true
				completedFileChan <- filePart
			} else if result == 401 {
				// bad auth token
				instance.b2Authorize()
				filePart.b2UploadPartURL()
				result = filePart.b2UploadPart()
			} else if result == 400 {
				// bad request
			} else if result == 999 {
				instance.b2Authorize()
				filePart.b2UploadPartURL()
				result = filePart.b2UploadPart()
			}
		}
	}
}

func folderMonitor(folder *Folders) {
	for {
		initialTime := time.Now()
		scanTime := ((time.Hour * time.Duration(folder.Hour)) + (time.Minute * time.Duration(folder.Minute)))
		time.Sleep(scanTime)
		var listFiles []string
		listFiles = getFiles(folder.RootFolder)
		for _, file := range listFiles {
			fileStat, _ := os.Stat(folder.RootFolder + "/" + file)
			fileTime := fileStat.ModTime()
			newFile := fileTime.After(initialTime)
			if newFile {
				var newFile File
				newFile.RootPath = folder.RootFolder
				newFile.B2Path = folder.B2Folder
				newFile.FilePath = file[1:]
				newFile.BucketID = folder.BucketID
				getSHAChan <- newFile
			}
		}
	}
}

func queueMonitor() {
	for {
		time.Sleep(time.Minute * 5)
		if len(fileCompleteQueue.Files) == 0 {
			exitChan <- true
		}
	}
}
