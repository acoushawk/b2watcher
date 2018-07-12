package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

type Config struct {
	AccountID       string     `yaml:"account_id"`
	AppKey          string     `yaml:"app_key"`
	ConnConnections int        `yaml:"con_connections"`
	API             APIConfig  `yaml:"api"`
	LogDir          string     `yaml:"log_dir"`
	Folders         []*Folders `yaml:"folders"`
}

type Folders struct {
	BucketID   string `yaml:"bucket_id"`
	B2Folder   string `yaml:"b2_folder"`
	RootFolder string `yaml:"root_folder"`
	Monitor    bool   `yaml:"monitor"`
	Hour       int    `yaml:"hour"`
	Minute     int    `yaml:"minute"`
	DeleteFile bool   `yaml:"delete_after_upload"`
	b2Files    B2ListFiles
}

type FileQueue struct {
	sync.Mutex
	Files []File
}

type File struct {
	FilePath  string
	B2Path    string
	SHA       string
	RootPath  string
	FileSize  int64
	FileID    string `json:"fileId"`
	BucketID  string `json:"bucketId"`
	Parts     []FilePart
	UploadURL B2UploadURL
}

type FilePart struct {
	Path         string
	ParentFileID string
	Number       int64
	SHA          string
	ChunkSize    int64
	Complete     bool
	URL          string
	AuthToken    string
	FileName     string
}

var configFilePath *string
var config Config

func main() {
	getFlags()
	config.parseConfig()
	instance.b2Authorize()
	for _, folder := range config.Folders {
		folder.b2Files.b2GetCurrentFiles(*folder)
	}
	go config.initialScan()
	daemon()
}

func getFiles(rootPath string) []string {
	var fileList []string
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		log.Fatal("Bad Folder location, please check config for ", rootPath)
	}
	err := filepath.Walk(rootPath, func(path string, f os.FileInfo, err error) error {
		if !(f.IsDir()) {
			fileList = append(fileList, strings.Replace(path, rootPath, "", 1))
		}
		return nil
	})
	if err != nil {
		log.Fatal("Error reading directory of files")
	}
	return fileList
}

func getFlags() {
	configFilePath = flag.String("config", "", "Specify the location to the configuration file to use")
	flag.Parse()
	// Let's make sure we have a config file present and verify it finds the file.
	if *configFilePath == "" {
		fmt.Println("Please supply a config file")
		os.Exit(1)
	} else if _, err := os.Stat(*configFilePath); os.IsNotExist(err) {
		fmt.Println("Config file not found")
		os.Exit(1)
	}
}

func (c *Config) initialScan() {
	for _, folder := range c.Folders {
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
				getSHAChan <- newFile
			}
		}
	}
}

func (c *Config) parseConfig() {
	configFile, _ := ioutil.ReadFile(*configFilePath)
	config.API.BindIP = "0.0.0.0"
	config.API.Port = "8000"
	err := yaml.UnmarshalStrict(configFile, &config)
	if err != nil {
		fmt.Println("There was an error reading the config file. Error was ", err)
		os.Exit(1)
	}
	if config.ConnConnections == 0 {
		config.ConnConnections = 4
	}
	for _, folder := range config.Folders {
		if (folder.Minute == 0) && (folder.Hour == 0) {
			folder.Minute = 30
		}
	}
}
