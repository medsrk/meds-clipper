package clipboard

import (
	"context"
	"golang.design/x/clipboard"
	"log"
)

// StartClipboardWatcher starts watching the clipboard for changes and triggers the provided callback with the new content.
func StartClipboardWatcher(ctx context.Context, onChange func(content string)) error {
	if err := clipboard.Init(); err != nil {
		return err
	}

	// Watch clipboard changes for text format.
	ch := clipboard.Watch(ctx, clipboard.FmtText)
	go func() {
		for data := range ch {
			// Since data is a slice of bytes, convert it to a string for text content.
			// This assumes the content is plain text.
			text := string(data)
			onChange(text)
		}
	}()
	log.Println("Clipboard watcher started")

	return nil
}
