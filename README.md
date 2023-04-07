# Torrent Pauser/Resumer based on Disk Space Left

This Go script pauses and resumes the torrent in qBittorrent based on the disk space left on the download directory. If the disk space falls below the specified margin, the script will pause the torrents to avoid filling up the disk. When the disk space increases above the margin, the script resumes the paused torrents.

# Configuration

The script requires a configuration file in YAML format to be provided with the following parameters:

- download_dir: the path to the download directory for qBittorrent. This is where the downloaded files will be saved.
- log_file_path: the path to the log file. The name of the log file will be "qbit.log".
- space_margin: the margin of disk space to be left in GB. When the disk space falls below this margin, the script pauses the torrents.
- qbittorrent_url: the URL of the qBittorrent web interface.
- qbittorrent_user: the username for the qBittorrent web interface.
- qbittorrent_password: the password for the qBittorrent web interface.
- skip_force_resume: if set to true, the script will not pause the forced resumed torrents. This is useful if you want to manually resume the torrents after the script has paused them.
- autoresume: if set to true, the script will automatically resume the torrents when the disk space increases above the margin. If set to false, the script will only pause the torrents when the disk space falls below the margin.

Here's an example configuration file:

```yaml
download_dir: "./downloads/"
log_file_path: "./qbit.log"
space_margin: 250
qbittorrent_url: "https://qbittorrent.host.com/"
qbittorrent_user: "username"
qbittorrent_password: "password"
skip_force_resume: true
autoresume: true
```

# Installation

To install download one of the releases from the releases page. Extract the archive and run the following command:

```bash
qbt --config /path/to/config.yml
```

# License

This project is licensed under the MIT License. See the LICENSE file for details.
