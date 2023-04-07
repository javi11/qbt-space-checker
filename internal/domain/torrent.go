package domain

import (
	qbittorrent "github.com/l3uddz/go-qbt"
	"github.com/l3uddz/go-qbt/pkg/model"
)

type Torrent struct {
	*model.Torrent
	amountLeft int
}

func (t *Torrent) GetAmountLeft() int {
	t.amountLeft = (100 - int(t.Progress)) / 100 * t.Size
	return t.amountLeft
}

func (t *Torrent) Pause(client *qbittorrent.Client) error {
	// Pause torrent
	return client.Torrent.StopTorrents([]string{t.Hash})
}

func (t *Torrent) Resume(client *qbittorrent.Client) error {
	// Resume torrent
	if err := client.Torrent.ResumeTorrents([]string{t.Hash}); err != nil {
		return err
	}

	return client.Torrent.ReannounceTorrents([]string{t.Hash})
}
