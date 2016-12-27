package url

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/pquerna/cachecontrol"
	"gopkg.in/redis.v5"

	"crypto/md5"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	DocumentNS string = "documentUrl:"
)

func HashMD5(s string) string {
	hasher := md5.New()
	hasher.Write([]byte(s))

	return hex.EncodeToString(hasher.Sum(nil))
}

func Document(r *redis.Client, urlStr string) (*goquery.Document, error) {
	hash := HashMD5(urlStr)
	key := DocumentNS + hash

	var doc *goquery.Document

	if !r.Exists(key).Val() {

		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return nil, err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if res.StatusCode != 200 {
			return nil, errors.New("Invalid response from remote document. Aborting.")
		}

		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		reasons, expires, _ := cachecontrol.CachableResponse(req, res, cachecontrol.Options{})
		expiration := 43200

		if len(reasons) == 0 {
			t := expires.Unix() - time.Now().Unix()

			if t > 0 {
				expiration = int(t)
			}
		}

		if expiration < 43200 {
			expiration = 43200
		}

		doc, err = goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, err
		}

		err = r.Set(key, string(body), time.Second*time.Duration(int(expiration))).Err()
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
