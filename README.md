# serps

[![Circle CI](https://img.shields.io/circleci/project/crackcomm/serps.svg)](https://circleci.com/gh/crackcomm/serps)

Consumes [google search](https://github.com/crackcomm/go-google-search) results from NSQ and stores in RethinkDB.

It also schedules a [crawl](https://github.com/crackcomm/crawl) of every page in search results.

## Usage

```
$ serps \
      --nsq-addr localhost:4150 \
      --nsqlookup-addr localhost:4161 \
      --nsq-topic google_results \
      --rethink-db default \
      --rethink-table serps \
      --rethink-addr localhost:28015 \
      --crawl-topic crawl \
      --crawl-callback github.com/crackcomm/…/spider.Example
```

## License

                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

## Authors

* [Łukasz Kurowski](https://github.com/crackcomm)
