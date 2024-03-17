package store

type Store interface {
	SaveClipboardItem(item string) error
	GetClipboardHistory() ([]string, error)
}
