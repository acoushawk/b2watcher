package main

import "fmt"

func (q *FileQueue) addFile(file File) {
	q.Lock()
	defer q.Unlock()
	q.Files = append(q.Files, file)
}

func (q *FileQueue) updateFile(filePart FilePart) {
	q.Lock()
	defer q.Unlock()
	for fileInt, file := range q.Files {
		if file.FileID == filePart.ParentFileID {
			for partInt, part := range file.Parts {
				if part.Number == filePart.Number {
					q.Files[fileInt].Parts[partInt].Complete = true
				}
			}
		}
	}
}

func (q *FileQueue) removeFile(file File) {
	q.Lock()
	defer q.Unlock()
	for i, queuefile := range q.Files {
		if queuefile.FilePath == file.FilePath {
			fmt.Println("removing slice ", i)
			fmt.Println("total slices ", len(q.Files))
			q.Files = append(q.Files[:i], q.Files[i+1:]...)
		}
	}
}

func (q *FileQueue) clearQueue() {
	for _, f := range q.Files {
		q.removeFile(f)
	}
}
