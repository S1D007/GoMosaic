package modules

import (
	"fmt"
	"mosaic/service"

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

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					service.OverlayImages(event.Name, gridCellFolder, outputFolder, opacity)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Error:", err)
			}
		}
	}()

	err = watcher.Add(inputFolder)
	if err != nil {
		fmt.Println("Error adding watcher:", err)
		return
	}
	<-done
}
