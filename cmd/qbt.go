package cmd

import (
	"flag"
	"log"
	"qbty-space-checker/internal/config"
	"qbty-space-checker/internal/domain"
	"qbty-space-checker/internal/logger"
	"qbty-space-checker/internal/torrentutils"
	"strings"

	qbittorrent "github.com/l3uddz/go-qbt"
	"github.com/l3uddz/go-qbt/pkg/model"
	"github.com/sirupsen/logrus"
)

func Manage() {
	// Setup logging and make things happen
	configFile := flag.String("config", "./config.yaml", "Manage Qbittorrent")
	flag.Parse()

	// Load the config file
	conf, err := config.Load(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	log, err := logger.New(conf.LogFilePath)
	if err != nil {
		panic(err)
	}

	log.Info("starting")

	run(conf, log)

	log.Info("shutdown")
}

func run(conf config.Config, log *logrus.Logger) {
	// Instantiate a Client using the appropriate WebUI configuration
	qb := qbittorrent.NewClient(strings.TrimSuffix(conf.QBittorrentURL, "/"), log)
	err := qb.Login(conf.QBittorrentUser, conf.QBittorrentPass)
	if err != nil {
		log.Fatal("Error on login. ", err)
	}

	log.Infof("connected to %s", conf.QBittorrentURL)

	// Gather information on the torrents. We manage tagging,
	// re-announce and identifying incomplete torrents here.
	count := 0
	paused := 0
	var incomplete []*domain.Torrent

	options := &model.GetTorrentListOptions{
		Filter: "all",
	}
	torrents, err := qb.Torrent.GetList(options)
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range torrents {
		// counters
		count++

		// Check for incomplete torrents
		// The double check here is in case a completed torrent has lost
		// its progress. We don't want to manually interfere with that!
		if t.Progress < 1 {
			torrent := &domain.Torrent{
				Torrent: t,
			}
			incomplete = append(incomplete, torrent)
		}

		// Count pause torrents but perform no extra checks on them
		// Specifically we don't want to try and reannounce paused torrents!
		if t.State == "pausedDL" {
			paused++
			continue
		}

		// Some other states to ignore
		if t.State == "checkingDL" || t.State == "checkingUP" {
			log.Printf("skipping %s - %s: %s [%s %.2f]\n", t.Category, t.Hash, t.Name, t.State, t.Progress)
			continue
		}

		trackers, err := qb.Torrent.GetTrackers(t.Hash)
		if err != nil {
			log.Fatal(err)
		}

		// If there is no tracker then try re-announcing...
		if len(trackers) == 0 {
			log.Printf("reannounce %s - %s: %s [%s]\n", t.Category, t.Hash, t.Name, t.State)
			err = qb.Torrent.ReannounceTorrents([]string{t.Hash})
			if err != nil {
				log.Println(err)
			}
			continue
		}
	}

	// Check incomplete torrents aren't going to blow disk space
	torrentutils.CheckAllIncomplete(incomplete, qb, conf, log)

	qb.Logout()
	log.Infof("processed %d torrents, %d paused", count, paused)
	log.Info("done")
}
