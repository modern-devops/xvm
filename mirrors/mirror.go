package mirrors

const Latest = "latest"

type Detail struct {
	MirrorBase string `json:"base"`
}

type Mirror interface {
	GetURL(v string) (string, error)
	GetLatestURL() (string, error)
	Detail() *Detail
}
