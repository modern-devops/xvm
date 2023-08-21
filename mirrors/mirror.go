package mirrors

const Latest = "latest"

type Mirror interface {
	GetURL(v string) (string, error)
	Versions() ([]string, error)
}
