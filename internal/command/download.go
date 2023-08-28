package command

type DownloadArguments struct {
	Artist string
	Album  string
	Track  string
}

func (args *DownloadArguments) IsValid() bool {
	return args.Artist != ""
}
