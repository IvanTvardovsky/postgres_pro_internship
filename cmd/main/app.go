package main

import (
	"database/sql"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"postgres_internship/internal/config"
	"postgres_internship/internal/utils"
	"postgres_internship/pkg/database"
	"sync"
	"syscall"
	"time"
)

func main() {
	cfg := config.GetConfig()
	db := database.Init(cfg)

	stopCh := make(chan bool)
	var wg sync.WaitGroup
	wg.Add(len(cfg.Directories))

	for _, directory := range cfg.Directories {
		go func(dir config.DirectoryConfig) {
			defer wg.Done()

			err := watchDirectory(dir.Path, dir.Commands, stopCh, db)
			if err != nil {
				log.Printf("Ошибка при установке наблюдателя для директории %s: %v", dir.Path, err)
			}
		}(directory)
	}

	waitForInterrupt(stopCh)

	wg.Wait()

	fmt.Println("Finished")
}

func watchDirectory(path string, commands []string, stopCh chan bool, db *sql.DB) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	defer watcher.Close()

	err = filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fileInfo.Mode().IsDir() {
			err = watcher.Add(filePath)
			if err != nil {
				log.Printf("Ошибка при добавлении директории %s в вотчер: %v", filePath, err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	done := make(chan bool)

	go func() {
		var actionNeeded bool
		var lastActionTime time.Time
		var fileName string
		firstChange := true

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					done <- true
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					fileName = event.Name
					if firstChange {
						firstChange = false
						actionNeeded = true
					} else {
						actionNeeded = true
						lastActionTime = time.Now()
					}
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					if fileInfo, err := os.Stat(event.Name); err == nil && fileInfo.IsDir() {
						err = watcher.Add(event.Name)
						if err != nil {
							log.Printf("Ошибка при добавлении директории %s в вотчер: %v", event.Name, err)
						}
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Ошибка в watcher:", err)
			}

			if actionNeeded && (firstChange || time.Since(lastActionTime) > 100*time.Millisecond) {

				date := utils.GetDate()

				_, err = db.Exec("INSERT INTO changes (file, date) VALUES ($1, $2)",
					fileName, date)
				if err != nil {
					log.Println("Ошибка в базе данных:", err)
				}

				var changeID int
				query := "SELECT id FROM changes WHERE file = $1 AND date = $2"
				err := db.QueryRow(query, fileName, date).Scan(&changeID)
				if err != nil {
					log.Println("Ошибка в базе данных:", err)
				}

				fmt.Printf("Файл изменен: %s\n", fileName)
				actionNeeded = false
				lastActionTime = time.Time{}

				for _, cmd := range commands {

					_, err = db.Exec("INSERT INTO executed_commands (command, date, change_id) VALUES ($1, $2, $3)",
						fileName, date, changeID)
					if err != nil {
						log.Println("Ошибка в базе данных:", err)
					}

					fmt.Printf("Executing %s\n", cmd)
					execution := exec.Command("cmd.exe", "/C", cmd)
					_, err := execution.Output()
					if err != nil {
						log.Printf("Error while executing command: %s\n%s", cmd, err)
						break
					}
				}
			}
		}
	}()

	go func() {
		<-stopCh
		watcher.Close()
		done <- true
	}()

	<-done

	return nil
}

func waitForInterrupt(stopCh chan bool) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	fmt.Println("Stopping...")

	close(stopCh)
}
