package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type B2Instance struct {
	AccountID   string `json:"accountId"`
	APIURL      string `json:"apiUrl"`
	AuthToken   string `json:"authorizationToken"`
	DownloadURL string `json:"downloadUrl"`
	RecPartSize int64  `json:"recommendedPartSize"`
}

type B2ListFiles struct {
	Files []struct {
		Action          string `json:"action"`
		ContentLength   int    `json:"contentLength"`
		FileID          string `json:"fileId"`
		FileName        string `json:"fileName"`
		Size            int    `json:"size"`
		UploadTimestamp int64  `json:"uploadTimestamp"`
	} `json:"files"`
	NextFileName string `json:"nextFileName"`
}

type B2ListFilesReq struct {
	BucketID      string `json:"bucketId"`
	StartFileName string `json:"startFileName"`
	MaxFileCount  int    `json:"maxFileCount"`
	Prefix        string `json:"prefix"`
}

type B2StartLargeFile struct {
	BucketID    string     `json:"bucketId"`
	FileName    string     `json:"fileName"`
	ContentType string     `json:"contentType"`
	FileInfo    B2FileInfo `json:"fileInfo"`
}

type B2FileInfo struct {
	LargeFileSHA string `json:"large_file_sha1"`
}

type B2UploadURL struct {
	UploadURL          string `json:"uploadUrl"`
	AuthorizationToken string `json:"authorizationToken"`
}

type B2FinishLargeFile struct {
	FileID        string   `json:"fileId"`
	PartSha1Array []string `json:"partSha1Array"`
}

var instance B2Instance

func (b *B2Instance) b2Authorize() {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.backblaze.com/b2api/v1/b2_authorize_account", nil)
	authHeader := base64.StdEncoding.EncodeToString([]byte(config.AccountID + ":" + config.AppKey))
	req.Header.Add("Authorization", "Basic "+string(authHeader))
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Backblaze returned error when authorizing. Error is ", err)
	}
	json.NewDecoder(resp.Body).Decode(&b)
}

func (f *B2ListFiles) b2GetCurrentFiles(folder Folders) {
	var tempListFiles B2ListFiles
	fileReq := B2ListFilesReq{
		BucketID:      folder.BucketID,
		StartFileName: f.NextFileName,
		MaxFileCount:  10000,
		Prefix:        folder.B2Folder,
	}
	body, _ := json.Marshal(fileReq)
	client := &http.Client{}
	url := instance.APIURL + "/b2api/v1/b2_list_file_names"
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", instance.AuthToken)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	json.NewDecoder(resp.Body).Decode(&tempListFiles)
	for _, file := range tempListFiles.Files {
		f.Files = append(f.Files, file)
	}
	f.NextFileName = tempListFiles.NextFileName
	if f.NextFileName != "" {
		f.b2GetCurrentFiles(folder)
	}
}

func (f *File) b2StartLargeFile() {
	var largeFileInfo B2StartLargeFile
	b2FileName, _ := url.Parse(f.B2Path + "/" + f.FilePath)
	largeFileInfo = B2StartLargeFile{
		BucketID:    f.BucketID,
		FileName:    b2FileName.String(),
		ContentType: "b2/x-auto",
		FileInfo: B2FileInfo{
			LargeFileSHA: f.SHA,
		},
	}
	body, _ := json.Marshal(largeFileInfo)
	client := &http.Client{}
	url := instance.APIURL + "/b2api/v1/b2_start_large_file"
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", instance.AuthToken)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	json.NewDecoder(resp.Body).Decode(&f)
}

func (f *File) b2UploadPartURL() {
	var uploadURL B2UploadURL
	fileID := map[string]string{"fileId": f.FileID}
	body, _ := json.Marshal(fileID)
	client := &http.Client{}
	url := instance.APIURL + "/b2api/v1/b2_get_upload_part_url"
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", instance.AuthToken)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	json.NewDecoder(resp.Body).Decode(&uploadURL)

	f.UploadURL = append(f.UploadURL, uploadURL)

}

func (f *FilePart) b2UploadPart() int {
	var bufferUp []byte
	log.Println("Uploading part ", f.Number, " of file ", f.Path)
	bufferUp = make([]byte, f.ChunkSize)
	openFile, _ := os.Open(f.Path)
	seek := ((f.Number - 1) * instance.RecPartSize)
	openFile.Seek(seek, 0)
	// go func() {
	openFile.Read(bufferUp)
	// }()
	url := f.URL
	req, _ := http.NewRequest("POST", url, bytes.NewReader(bufferUp))
	req.Header.Set("Authorization", f.AuthToken)
	req.Header.Add("X-Bz-Part-Number", strconv.Itoa(int(f.Number)))
	// req.Header.Add("Content-Length", strconv.Itoa(int(f.ChunkSize)))
	req.ContentLength = f.ChunkSize
	req.Header.Add("X-Bz-Content-Sha1", f.SHA)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error in send")
		fmt.Println(err)
		openFile.Close()
		return resp.StatusCode
	}
	resp.Body.Close()
	openFile.Close()
	return resp.StatusCode
}

func (f *File) b2FinishLargeFile() {
	var finishLargeFile B2FinishLargeFile
	var shaArray []string
	for _, part := range f.Parts {
		shaArray = append(shaArray, part.SHA)
	}
	finishLargeFile = B2FinishLargeFile{
		FileID:        f.FileID,
		PartSha1Array: shaArray,
	}
	body, _ := json.Marshal(finishLargeFile)
	url := instance.APIURL + "/b2api/v1/b2_finish_large_file"
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", instance.AuthToken)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	result, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(result))
}

func (f *File) b2UploadURL() {
	var uploadURL B2UploadURL
	bucketID := map[string]string{"bucketId": f.BucketID}
	body, _ := json.Marshal(bucketID)
	url := instance.APIURL + "/b2api/v1/b2_get_upload_url"
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Authorization", instance.AuthToken)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error")
	}
	json.NewDecoder(resp.Body).Decode(&uploadURL)
	f.UploadURL = append(f.UploadURL, uploadURL)
}

func (f *FilePart) b2UploadFile() int {
	var bufferUp []byte
	log.Println("Starting upload of file ", f.FileName)
	bufferUp = make([]byte, f.ChunkSize)
	openFile, _ := os.Open(f.Path)
	openFile.Read(bufferUp)
	url := f.URL
	req, _ := http.NewRequest("POST", url, bytes.NewReader(bufferUp))
	req.Header.Set("Authorization", f.AuthToken)
	req.Header.Add("X-Bz-File-Name", f.FileName)
	req.Header.Add("Content-Type", "b2/x-auto")
	req.Header.Add("X-Bz-Content-Sha1", f.SHA)
	req.ContentLength = f.ChunkSize
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("There was an error sending part ", f.Number, " of file ", f.FileName)
		openFile.Close()
		return resp.StatusCode
	}
	openFile.Close()
	return resp.StatusCode
}
