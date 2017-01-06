# Spider service

The spider service allows to track & manipulate sitemaps, configure datasets based on css selectors & apply selectors to generate groups of datasets for URLs. 

## Getting started

1. Install Consul

	Consul is the default registry/discovery for go-micro apps. It's however pluggable.
	[https://www.consul.io/intro/getting-started/install.html](https://www.consul.io/intro/getting-started/install.html)

2. Run Consul
	```
	$ consul agent -server -bootstrap-expect 1 -data-dir /tmp/consul
	```

3. Start a postgres database

4. Download and start the service

	```shell
	go get github.com/suicidegang/spider-srv
	spider-srv --pgsql_url="host=127.0.0.1 user=fernandez14 dbname=comparateca sslmode=disable"
	```

## The API
Spider server implements the following RPC Methods

Spider
- TrackSitemap
- FetchDataset
- PrepareDatasets
- FetchPages


### Spider.TrackSitemap
```shell
$ micro query sg.micro.srv.spider Spider.TrackSitemap '{
    "from": "https://www.whistleout.com.mx/CellPhones",
    "name": "Whistleout cellphones #4",
    "depth": 4,
    "patterns": {
        "home": "^https://www\\.whistleout\\.com\\.mx/CellPhones/?$",
        "carrier": "^https://www\\.whistleout\\.com\\.mx/CellPhones/Carriers/([[:alnum:]\\-]+)$",
        "phone": "^https://www\\.whistleout\\.com\\.mx/CellPhones/Phones/([[:alnum:]\\-]+)/([[:alnum:]\\-]+)$",
        "phone-plans": "^https://www\\.whistleout\\.com\\.mx/CellPhones/Carriers/([[:alnum:]\\-]+)/Phones/([[:alnum:]\\-]+)/([[:alnum:]\\-]+)$",
        "plan": "^https://www\\.whistleout\\.com\\.mx/CellPhones/Carriers/([[:alnum:]\\-]+)/([[:alnum:]\\-]+)/([[:alnum:]\\-]+)(\\??[[:alnum:]\\=\\&_]+)$"
    }
}'
```

### Spider.FetchDataset
```shell
$ micro query sg.micro.srv.spider Spider.FetchDataset '{
    "id": 1,
    "url_id": 14
}'
```

### Spider.PrepareDatasets
```shell
$ micro query sg.micro.srv.spider Spider.PrepareDatasets '{
    "selector_id": 1,
    "group": "phone"
}'
```

### Spider.FetchPages
```shell
$ micro query sg.micro.srv.spider Spider.PrepareDatasets '{
    "search": "vas a volar iphone 6",
    "group": "plan"
}'
```

## Selector example
```json
{
	"name": {"use": "text", "value": "h1 span", "filters": ["trim-space"]},
	"screen": {"use": "text", "value": "#phone div[data-gallery] .col-xs-16 div.bor-t-1", "filters": ["trim-space"]},
	"os": {"use": "text", "value": "#phone div[data-gallery] .col-xs-16 div.bor-t-1", "i": 1, "filters": ["trim-space"]},
	"camera": {"use": "text", "value": "#phone div[data-gallery] .col-xs-16 div.bor-t-1", "i": 2, "filters": ["trim-space"]}, 
	"storage": {"use": "text", "value": "#phone div[data-gallery] .col-xs-16 div.bor-t-1", "i": 3, "filters": ["trim-space"]},
}
```
