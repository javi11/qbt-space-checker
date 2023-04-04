package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"syscall"
	"time"

	"github.com/superturkey650/go-qbittorrent/qbt"
)

const secondsInDay = 60 * 60 * 24

type params struct {
	QBittorrentURL   string `json:"qbittorrent_url"`
	QBittorrentUser  string `json:"qbittorrent_user"`
	QBittorrentPass  string `json:"qbittorrent_pass"`
	DownloadLocation string `json:"download_location"`
	SpaceMargin      int64  `json:"space_margin"`
	TrackerRegex     string `json:"tracker_regex"`
	AutoResume       bool   `json:"autoresume"`
	LogFilePath      string `json:"log_file_path"`
}

func main() {
	// Setup logging and make things happen
	configFile := flag.String("config", "./config.json", "Manage Qbittorrent")
	autoResume := flag.Bool("auto-resume", false, "Auto resume")
	noAutoResume := flag.Bool("no-auto-resume", false, "No auto resume")
	flag.Parse()

	if *noAutoResume {
		*autoResume = false
	}

	// Load the config file
	params, err := loadParams(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	logFile, err := createLogFile(params.LogFilePath)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(logFile, "[INFO] ", log.LstdFlags)
	logger.Println("startup")

	doWork(params)

	logger.Println("shutdown")
}

func createLogFile(logFilePath string) (*os.File, error) {

	// Create the log file if it doesn't exist
	logFile := logFilePath + "/qbit.log"

	// Check if the file exists
	return os.OpenFile(logFile, os.O_RDONLY|os.O_CREATE, 0666)
}

func doWork(params params) {
	// Instantiate a Client using the appropriate WebUI configuration
	qb := qbt.NewClient(params.QBittorrentURL)
	_, err := qb.Login(params.QBittorrentUser, params.QBittorrentPass)
	if err != nil {
		log.Fatal(err)
	}
	defer qb.Logout()

	// Gather information on the torrents. We manage tagging,
	// re-announce and identifying incomplete torrents here.
	count := 0
	paused := 0
	var incomplete []*torrent.Torrent
	amountLeft := int64(0)

	torrents, err := qbtClient.Torrents()
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range torrents {
		// counters
		count++

		// Check for incomplete torrents
		// The double check here is in case a completed torrent has lost
		// its progress. We don't want to manually interfere with that!
		if t.BytesCompleted() != t.Info().TotalLength() && t.Completion().IsZero() {
			incomplete = append(incomplete, t)
			amountLeft += t.BytesMissing()
		}

		// Count pause torrents but perform no extra checks on them
		// Specifically we don't want to try and reannounce paused torrents!
		if t.Paused() {
			paused++
			continue
		}

		// Some other states to ignore
		if t.Checking() {
			log.Printf("skipping %s - %s: %s [%s %.2f]\n", t.Category(), t.InfoHash().HexString()[26:], t.Name(), t.Status(), t.Progress())
			continue
		}

		// If there is no tracker then try re-announcing...
		if len(t.Trackers()) == 0 {
			log.Printf("reannounce %s - %s: %s [%s]\n", t.Category(), t.InfoHash().HexString()[26:], t.Name(), t.Status())
			err = t.Announce()
			if err != nil {
				log.Println(err)
			}
			continue
		}
	}

	// Check incomplete torrents aren't going to blow disk space
	checkAllIncomplete(incomplete, qbtClient, params)
}

func checkAllIncomplete(incomplete []Torrent, qbtClient *QbtClient, params map[string]interface{}) {
	totalActive := 0
	totalPaused := 0
	var paused []Torrent
	var active []Torrent

	for _, inc := range incomplete {
		if inc.State == "pausedDL" {
			if inc.CompletionOn <= 0 {
				totalPaused += inc.AmountLeft
				paused = append(paused, inc)
			}
		} else {
			totalActive += inc.AmountLeft
			active = append(active, inc)
		}
	}

	// Get free space and adjust for the min free space (GB)
	free := freeSpace(params["download_dir"].(string))
	free = free - (params["min_free_gb"].(int) * 1024 * 1024 * 1024)

	// Optionally autoresume torrents if there is space available
	if params["autoresume"].(bool) {
		totalLeft := totalPaused + totalActive
		if totalLeft < free {
			// unpause everything
			for _, torrent := range paused {
				resume(torrent, qbtClient)
				totalPaused += torrent.AmountLeft
				totalActive += torrent.AmountLeft
			}
			return
		}

		// Otherwise do a piecemeal check to see if we can resume
		// any currently paused torrents
		if totalActive < free {
			// Check to see if we can resume anything paused, starting
			// with the torrents with the least remaining
			sort.Slice(paused, func(i, j int) bool {
				return paused[i].AmountLeft < paused[j].AmountLeft
			})
			for _, torrent := range paused {
				if (totalActive + torrent.AmountLeft) < free {
					resume(torrent, qbtClient)
					totalActive += torrent.AmountLeft
					totalPaused -= torrent.AmountLeft
				}
			}
			return
		}
	}

	// Finally we have more active than free so must pause some
	// Start with the ones with the most remaining
	sort.Slice(active, func(i, j int) bool {
		return active[i].AmountLeft > active[j].AmountLeft
	})
	for _, torrent := range active {
		pause(torrent, qbtClient)
		totalActive -= torrent.AmountLeft
		totalPaused += torrent.AmountLeft

		// We only need to pause enough to get under free space
		if totalActive < free {
			break
		}
	}
}

func loadParams(configFile string) (params, error) {
	// Load parameters from file
	f, err := os.Open(configFile)
	if err != nil {
		return params{}, err
	}
	defer f.Close()

	var p params
	err = json.NewDecoder(f).Decode(&p)
	if err != nil {
		return params{}, err
	}

	return p, nil
}

func sizeofFmt(num int64) string {
	// Convert number of bytes into human-readable format
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
	for _, unit := range units {
		if num < 1024 {
			return fmt.Sprintf("%3.1f%s", float64(num), unit)
		}
		num /= 1024
	}
	return fmt.Sprintf("%.1f%s", float64(num), "YiB")
}

func freeSpace(directory string) (int64, error) {
	// Get free space on download area
	var fs syscall.Statfs_t
	err := syscall.Statfs(directory, &fs)
	if err != nil {
		return 0, err
	}

	freeBytes := fs.Bavail * uint64(fs.Bsize)
	return int64(freeBytes), nil
}

func pause(torrent *qb.Torrent, client *qb.Client) error {
	// Pause torrent
	log.Printf("Pause - %s: %s", torrent.Hash[-6:], torrent.Name)
	return client.TorrentsPause([]string{torrent.Hash})
}

func resume(torrent *qb.Torrent, client *qb.Client) error {
	// Resume torrent
	log.Printf("Resume - %s: %s", torrent.Hash[-6:], torrent.Name)
	if err := client.TorrentsResume([]string{torrent.Hash}); err != nil {
		return err
	}

	// Reannounce just to make sure
	log.Printf("Reannounce - %s: %s", torrent.Hash[-6:], torrent.Name)
	return client.TorrentsReannounce([]string{torrent.Hash})
}

func contains(str string, regex string) bool {
	// Check if a string contains a regex
	matched, _ := regexp.MatchString(regex, str)
	return matched
}

func days(torrent *qb.Torrent) float64 {
	// Return age of torrent in days
	return time.Since(time.Unix(torrent.CompletionOn, 0)).Hours() / 24
}
