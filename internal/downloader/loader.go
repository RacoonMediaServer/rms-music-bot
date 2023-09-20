package downloader

type loader struct {
	db Database
}

func (l loader) ListTorrents() (map[string][][]byte, error) {
	list, err := l.db.LoadTorrents()
	if err != nil {
		return nil, err
	}

	var torrents [][]byte
	for _, t := range list {
		torrents = append(torrents, t.Bytes)
	}

	return map[string][][]byte{mainRoute: torrents}, nil
}
