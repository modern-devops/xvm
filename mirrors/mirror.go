package mirrors

const Latest = "latest"

type Mirror interface {
	GetURL(v string) (string, error)
	GetLatestURL() (string, error)
}
