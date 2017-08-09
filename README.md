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
    "from": "http://www.data.com/specs/",
    "name": "Data specs",
    "depth": 4,
    "strict": true, 
    "patterns": [
        {"name": "motorcycle", "matches": "^{manufacturer:slug}/{type:slug}/{year:num}/{series:slug}/{model:slug}.html"},
        {"name": "filtered-list", "matches": "^{list:slug}/{filter:slug}.html"},
        {"name": "list", "matches": "^{list:slug}.html"},
        {"name": "landing", "matches": "^$"}
    ]
}
```
### Spider.TrackURL
```shell
$ micro query sg.micro.srv.spider Spider.TrackURL '{
    "url": "https://comparateca.com/iphone-6",
    "group": "comparateca"
}'
```

micro query sg.micro.srv.spider Spider.TrackURL '{
    "url": "http://www.movilcelular.es/movil/motorola-moto-g5-plus-xt1681-32gb/3207",
    "group": "movilcelular"
}'

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
$ micro query sg.micro.srv.spider Spider.FetchPages '{
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
