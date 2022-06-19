package urlshorten

import (
	"math/big"
)

type UniqueIDGenerator func() (int64, error)

type UrlShortener struct {
	idGenerator UniqueIDGenerator
}

func (us *UrlShortener) Shorten(url string) (string, error) {
	id, err := us.idGenerator()
	if err != nil {
		return "", err
	}

	return toBase62(id), nil
}

func toBase62(n int64) string {
	return big.NewInt(n).Text(62)
}
