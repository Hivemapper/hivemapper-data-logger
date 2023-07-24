package camera

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
)

type ImageHandler func(imageFileName string) error

type ImageFeed struct {
	handlers     []ImageHandler
	imagesFolder string
}

func NewImageFeed(imagesFolder string, handlers ...ImageHandler) *ImageFeed {
	return &ImageFeed{
		handlers:     handlers,
		imagesFolder: imagesFolder,
	}
}

func (f *ImageFeed) Run() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("NewWatcher failed: ", err)
	}

	err = watcher.Add(f.imagesFolder)
	if err != nil {
		return fmt.Errorf("adding folder to watcher %s: %s", f.imagesFolder, err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("watcher channel closed")
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				for _, handler := range f.handlers {
					err := handler(event.Name)
					if err != nil {
						return fmt.Errorf("calling handler: %w", err)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher channel closed")
			}
			return fmt.Errorf("watcher error: %w", err)
		}
	}

	return nil
}
