package url

import (
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/redis.v5"

	"crypto/md5"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	DocumentNS string = "documentUrl:"
)

func Document(r *redis.Client, urlStr string) (*goquery.Document, error) {
	hasher := md5.New()
	hasher.Write([]byte(urlStr))
	hash := hex.EncodeToString(hasher.Sum(nil))
	key := DocumentNS + hash

	var doc *goquery.Document

	if !r.Exists(key).Val() {
		res, err := http.Get(urlStr)
		if err != nil {
			return nil, err
		}

		if res.StatusCode != 200 {
			return nil, errors.New("Invalid response from remote document. Aborting.")
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		doc, err = goquery.NewDocumentFromResponse(res)
		if err != nil {
			return nil, err
		}

		err = r.Set(key, string(body), 0).Err()
		if err != nil {
			return nil, err
		}
	} else {
		body, err := r.Get(key).Result()
		if err != nil {
			return nil, err
		}

		reader := strings.NewReader(body)
		doc, err = goquery.NewDocumentFromReader(reader)
		if err != nil {
			return nil, err
		}
	}

	return doc, nil
}
