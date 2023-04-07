package torrentutils

import (
	"qbty-space-checker/internal/config"
	"qbty-space-checker/internal/domain"
	"qbty-space-checker/internal/utils"
	"sort"

	qbittorrent "github.com/l3uddz/go-qbt"
	"github.com/sirupsen/logrus"
)

func CheckAllIncomplete(incomplete []*domain.Torrent, qbtClient *qbittorrent.Client, conf config.Config, log *logrus.Logger) {
	totalActive := 0
	totalPaused := 0
	var paused []*domain.Torrent
	var active []*domain.Torrent

	for _, inc := range incomplete {
		if inc.State == "pausedDL" {
			if inc.Eta <= 0 {
				totalPaused += inc.Size
				paused = append(paused, inc)
			}
		} else {
			totalActive += inc.GetAmountLeft()
			active = append(active, inc)
		}
	}

	// Get free space and adjust for the min free space (GB)
	free, err := utils.FreeSpace(conf.DownloadLocation)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("free space on system (GB): %v", utils.ToGB(int(free)))
	free = free - (conf.SpaceMargin * 1024 * 1024 * 1024)
	log.Infof("space left (GB) to fill with torrents: %v", utils.ToGB(free))

	// Optionally autoresume torrents if there is space available
	if conf.AutoResume {
		totalLeft := totalPaused + totalActive
		if totalLeft < free {
			// unpause everything
			for _, torrent := range paused {
				log.Infof("resume/reannounce - %s: %s", torrent.Hash, torrent.Name)
				torrent.Resume(qbtClient)
				totalPaused += torrent.GetAmountLeft()
				totalActive += torrent.GetAmountLeft()
			}
			return
		}

		// Otherwise do a piecemeal check to see if we can resume
		// any currently paused torrents
		if totalActive < free {
			// Check to see if we can resume anything paused, starting
			// with the torrents with the least remaining
			sort.Slice(paused, func(i, j int) bool {
				return paused[i].GetAmountLeft() < paused[j].GetAmountLeft()
			})
			for _, torrent := range paused {
				if (totalActive + torrent.GetAmountLeft()) < free {
					log.Infof("resume/reannounce - %s: %s", torrent.Hash, torrent.Name)
					torrent.Resume(qbtClient)
					totalActive += torrent.GetAmountLeft()
					totalPaused -= torrent.GetAmountLeft()
				}
			}
			return
		}
	}

	// Finally we have more active than free so must pause some
	// Start with the ones with the most remaining
	sort.Slice(active, func(i, j int) bool {
		return active[i].GetAmountLeft() > active[j].GetAmountLeft()
	})

	for _, torrent := range active {
		if torrent.ForceStart && conf.SkipForceResume {
			log.Infof("skipping %s - %s: %s [%s %.2f] (force resume is enabled)", torrent.Category, torrent.Hash, torrent.Name, torrent.State, torrent.Progress)
			continue
		}
		log.Infof("pause - %s: %s", torrent.Hash, torrent.Name)
		torrent.Pause(qbtClient)
		totalActive -= torrent.GetAmountLeft()
		totalPaused += torrent.GetAmountLeft()

		// We only need to pause enough to get under free space
		if totalActive < free {
			break
		}
	}
}
