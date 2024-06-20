package modules

import (
	"fmt"
	"mosaic/service"
	"time"

	"github.com/fsnotify/fsnotify"
)


func MonitorInputFolder(gridCellFolder, inputFolder, outputFolder string, opacity float64) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error creating watcher:", err)
		return
	}
	defer watcher.Close()

	done := make(chan bool)
	eventQueue := make(chan string, 1000)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					eventQueue <- event.Name
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Error:", err)
			}
		}
	}()

	go func() {
		for fileName := range eventQueue {
			service.OverlayImages(fileName, gridCellFolder, outputFolder, opacity)
			time.Sleep(100 * time.Millisecond)
		}
	}()
	// go func() {
	// 	for {
	// 		select {
	// 		case event, ok := <-watcher.Events:
	// 			if !ok {
	// 				return
	// 			}
	// 			if event.Op&fsnotify.Create == fsnotify.Create {
	// 				service.OverlayImages(event.Name, gridCellFolder, outputFolder, opacity)
	// 			}
	// 		case err, ok := <-watcher.Errors:
	// 			if !ok {
	// 				return
	// 			}
	// 			fmt.Println("Error:", err)
	// 		}
	// 	}
	// }()

	err = watcher.Add(inputFolder)
	if err != nil {
		fmt.Println("Error adding watcher:", err)
		return
	}
	<-done
}
