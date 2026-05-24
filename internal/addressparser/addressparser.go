package addressparser

type Match struct {
	Name           string
	NormalizedName string
	Confidence     float64
}

type AddressParser struct{}

func New() *AddressParser {
	return &AddressParser{}
}
