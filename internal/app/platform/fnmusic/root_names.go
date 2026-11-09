package fnmusic

import (
	"strings"

	backend "netease-music-mygo/internal/platform/fnmusic"
)

var cloudStorageNames = map[int]string{
	1:  "Baidu Netdisk",
	2:  "Alibaba Cloud Drive",
	3:  "115 生活",
	4:  "Quark Drive",
	5:  "123pan",
	6:  "Microsoft OneDrive",
	7:  "Google Drive",
	8:  "Dropbox",
	9:  "Microsoft OneDrive for Business",
	10: "Microsoft OneDrive",
}

func rootPathKey(path string) string {
	return strings.TrimRight(strings.TrimSpace(path), "/")
}

func authorizedDirectoryName(dir backend.AuthorizedDirectory) string {
	if dir.StorageType != 2 {
		return ""
	}
	comment := strings.TrimSpace(dir.Comment)
	if cloud := cloudStorageNames[dir.CloudStorageType]; cloud != "" {
		if comment != "" {
			return cloud + " - " + comment
		}
		return cloud
	}
	account := strings.TrimSpace(dir.Username)
	if address := strings.TrimSpace(dir.Address); address != "" {
		if account != "" {
			account += "@"
		}
		account += address
	}
	switch {
	case comment != "" && account != "":
		return comment + " - " + account
	case comment != "":
		return comment
	case account != "":
		return account
	default:
		return strings.TrimSpace(dir.Uname)
	}
}

func rootNames(dirs []backend.AuthorizedDirectory) map[string]string {
	out := map[string]string{}
	for _, dir := range dirs {
		key := rootPathKey(dir.Path)
		if key == "" {
			continue
		}
		if name := authorizedDirectoryName(dir); name != "" {
			out[key] = name
		}
	}
	return out
}
