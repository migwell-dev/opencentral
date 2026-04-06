package main

import (
	"fmt"
	"log"
	"net/http"
	"opencentral/cron"
	"opencentral/db"
	"opencentral/src"
)

func main() {
	// create storage if doesnt exist
	src.InitStorage()
	db.InitDB()
	src.SyncFiles()
	cron.StartCronJobs()
	port := db.GetSetting("port", "8080")

	// health
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// routes
	http.HandleFunc("/api/files", src.ListFiles)
	http.HandleFunc("/api/files/trash", src.ListTrashFiles)
	http.HandleFunc("/api/file", src.ServeFile)
	http.HandleFunc("/api/files/star", src.StarFiles)
	http.HandleFunc("/api/upload", src.StorageLimitMiddleware(src.UploadFile))
	http.HandleFunc("/api/delete", src.DeleteFile)
	http.HandleFunc("/api/rename", src.RenameFile)
	http.HandleFunc("/api/move", src.MoveFile)
	http.HandleFunc("/api/mkdir", src.CreateDirectory)
	http.HandleFunc("/api/trash", src.TrashFile)
	http.HandleFunc("/api/restore", src.RestoreFile)
	http.HandleFunc("/api/storage/usage", src.GetStorageUsage)
	http.HandleFunc("/api/settings", src.GetSettings)
	http.HandleFunc("/api/settings/update", src.UpdateSetting)
	http.HandleFunc("/api/download/batch", src.BatchDownload)

	// serve files
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./storage"))))

	fmt.Println("OpenCentral running on http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
