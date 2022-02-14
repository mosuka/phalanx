# Wikipedia example

To experience Phalanx functionality, let's start Phalanx with [MinIO](https://min.io/) and [etcd](https://etcd.io/). 
This repository has a docker-compose.yml file. With it, you can easily launch MinIO and etcd on Docker.

```
% docker-compose up --force-recreate etcd etcdkeeper minio
```

Once the container has been started, you can check the MinIO and etcd data in your browser at the following URL.

- MinIO  
http://localhost:9001/dashboard

- ETCD Keeper  
http://localhost:8080/etcdkeeper/


## Start Phalanx with etcd metastore

Start the first node:

```
% ./bin/phalanx --index-metastore-uri=etcd://phalanx-metadata
```

Start the second node:

```
% ./bin/phalanx --index-metastore-uri=etcd://phalanx-metadata --bind-port=2001 --grpc-port=5001 --http-port=8001 --seed-addresses=localhost:2000
```

Start the third node:

```
% ./bin/phalanx --index-metastore-uri=etcd://phalanx-metadata --bind-port=2002 --grpc-port=5002 --http-port=8002 --seed-addresses=localhost:2000
```

Use the following command to create an index of 6 shards. If you need more nodes, start them in the same way as in the above command.


### Create index with MinIO and etcd

Use MinIO as index storage, and create a lock on etcd to avoid write conflicts.

```
% curl -XPUT -H 'Content-type: application/json' http://localhost:8000/v1/indexes/enwiki --data-binary '
{
	"index_uri": "minio://phalanx-indexes/enwiki",
	"lock_uri": "etcd://phalanx-locks/enwiki",
	"index_mapping": {
		"id": {
			"type": "numeric",
			"options": {
				"index": true,
				"store": true,
				"sortable": false,
				"aggregatable": false
			}
		},
		"url": {
			"type": "text",
			"options": {
				"index": true,
				"store": true,
				"term_positions": false,
				"highlight": false,
				"sortable": false,
				"aggregatable": true
			},
			"analyzer": {
				"char_filters": [
					{
						"name": "ascii_folding"
					},
					{
						"name": "unicode_normalize",
						"options": {
							"form": "NFKC"
						}
					}
				],
				"tokenizer": {
					"name": "single_token"
				},
				"token_filters": [
					{
						"name": "lower_case"
					}
				]
			}
		},
		"title": {
			"type": "text",
			"options": {
				"index": true,
				"store": true,
				"term_positions": true,
				"highlight": true,
				"sortable": false,
				"aggregatable": false
			},
			"analyzer": {
				"char_filters": [
					{
						"name": "ascii_folding"
					},
					{
						"name": "unicode_normalize",
						"options": {
							"form": "NFKC"
						}
					}
				],
				"tokenizer": {
					"name": "unicode"
				},
				"token_filters": [
					{
						"name": "lower_case"
					}
				]
			}
		},
		"text": {
			"type": "text",
			"options": {
				"index": true,
				"store": true,
				"term_positions": true,
				"highlight": true,
				"sortable": false,
				"aggregatable": false
			},
			"analyzer": {
				"char_filters": [
					{
						"name": "ascii_folding"
					},
					{
						"name": "unicode_normalize",
						"options": {
							"form": "NFKC"
						}
					}
				],
				"tokenizer": {
					"name": "unicode"
				},
				"token_filters": [
					{
						"name": "lower_case"
					}
				]
			}
		},
		"categories": {
			"type": "text",
			"options": {
				"index": true,
				"store": true,
				"term_positions": false,
				"highlight": false,
				"sortable": false,
				"aggregatable": true
			},
			"analyzer": {
				"char_filters": [
					{
						"name": "ascii_folding"
					},
					{
						"name": "unicode_normalize",
						"options": {
							"form": "NFKC"
						}
					}
				],
				"tokenizer": {
					"name": "single_token"
				},
				"token_filters": [
					{
						"name": "lower_case"
					}
				]
			}
		},
		"external_links": {
			"type": "text",
			"options": {
				"index": true,
				"store": true,
				"term_positions": false,
				"highlight": false,
				"sortable": false,
				"aggregatable": true
			},
			"analyzer": {
				"char_filters": [
					{
						"name": "ascii_folding"
					},
					{
						"name": "unicode_normalize",
						"options": {
							"form": "NFKC"
						}
					}
				],
				"tokenizer": {
					"name": "single_token"
				},
				"token_filters": [
					{
						"name": "lower_case"
					}
				]
			}
		},
		"links": {
			"type": "text",
			"options": {
				"index": true,
				"store": true,
				"term_positions": false,
				"highlight": false,
				"sortable": false,
				"aggregatable": true
			},
			"analyzer": {
				"char_filters": [
					{
						"name": "ascii_folding"
					},
					{
						"name": "unicode_normalize",
						"options": {
							"form": "NFKC"
						}
					}
				],
				"tokenizer": {
					"name": "single_token"
				},
				"token_filters": [
					{
						"name": "lower_case"
					}
				]
			}
		},
		"media": {
			"type": "text",
			"options": {
				"index": true,
				"store": true,
				"term_positions": false,
				"highlight": false,
				"sortable": false,
				"aggregatable": true
			},
			"analyzer": {
				"char_filters": [
					{
						"name": "ascii_folding"
					},
					{
						"name": "unicode_normalize",
						"options": {
							"form": "NFKC"
						}
					}
				],
				"tokenizer": {
					"name": "single_token"
				},
				"token_filters": [
					{
						"name": "lower_case"
					}
				]
			}
		},
		"redirect": {
			"type": "text",
			"options": {
				"index": true,
				"store": true,
				"term_positions": false,
				"highlight": false,
				"sortable": false,
				"aggregatable": true
			},
			"analyzer": {
				"char_filters": [
					{
						"name": "ascii_folding"
					},
					{
						"name": "unicode_normalize",
						"options": {
							"form": "NFKC"
						}
					}
				],
				"tokenizer": {
					"name": "single_token"
				},
				"token_filters": [
					{
						"name": "lower_case"
					}
				]
			}
		},
    "timestamp": {
      "type": "datetime",
      "options": {
        "index": true,
        "store": true,
        "sortable": true,
        "aggregatable": true
      }
    }
	},
	"num_shards": 6,
	"default_search_field": "_all",
	"default_analyzer": {
		"tokenizer": {
			"name": "unicode"
		},
		"token_filters": [
			{
				"name": "lower_case"
			}
		]
	}
}
'
```


### Cluster status

```
% curl -XGET http://localhost:8000/cluster | jq .
```

```json
{
  "indexer_assignment": {
    "example": {
      "shard-0kZFXhrZ": "node-8giCE4QA",
      "shard-FretswBz": "node-Rh6kmVO8",
      "shard-Lf7rlqwd": "node-Rh6kmVO8",
      "shard-WvUd7lWm": "node-Rh6kmVO8",
      "shard-f4f6jbCi": "node-Rh6kmVO8",
      "shard-kCPFctnO": "node-8giCE4QA"
    }
  },
  "indexes": {
    "example": {
      "index_lock_uri": "etcd://phalanx-locks/example",
      "index_mapping": {
        "categories": {
          "type": "text",
          "options": {
            "index": true,
            "store": true,
            "term_positions": false,
            "highlight": false,
            "sortable": false,
            "aggregatable": true
          },
          "analyzer": {
            "char_filters": [
              {
                "name": "ascii_folding",
                "options": null
              },
              {
                "name": "unicode_normalize",
                "options": {
                  "form": "NFKC"
                }
              }
            ],
            "tokenizer": {
              "name": "single_token",
              "options": null
            },
            "token_filters": [
              {
                "name": "lower_case",
                "options": null
              }
            ]
          }
        },
        "external_links": {
          "type": "text",
          "options": {
            "index": true,
            "store": true,
            "term_positions": false,
            "highlight": false,
            "sortable": false,
            "aggregatable": true
          },
          "analyzer": {
            "char_filters": [
              {
                "name": "ascii_folding",
                "options": null
              },
              {
                "name": "unicode_normalize",
                "options": {
                  "form": "NFKC"
                }
              }
            ],
            "tokenizer": {
              "name": "single_token",
              "options": null
            },
            "token_filters": [
              {
                "name": "lower_case",
                "options": null
              }
            ]
          }
        },
        "id": {
          "type": "numeric",
          "options": {
            "index": true,
            "store": true,
            "term_positions": false,
            "highlight": false,
            "sortable": false,
            "aggregatable": false
          },
          "analyzer": {
            "char_filters": null,
            "tokenizer": {
              "name": "",
              "options": null
            },
            "token_filters": null
          }
        },
        "links": {
          "type": "text",
          "options": {
            "index": true,
            "store": true,
            "term_positions": false,
            "highlight": false,
            "sortable": false,
            "aggregatable": true
          },
          "analyzer": {
            "char_filters": [
              {
                "name": "ascii_folding",
                "options": null
              },
              {
                "name": "unicode_normalize",
                "options": {
                  "form": "NFKC"
                }
              }
            ],
            "tokenizer": {
              "name": "single_token",
              "options": null
            },
            "token_filters": [
              {
                "name": "lower_case",
                "options": null
              }
            ]
          }
        },
        "media": {
          "type": "text",
          "options": {
            "index": true,
            "store": true,
            "term_positions": false,
            "highlight": false,
            "sortable": false,
            "aggregatable": true
          },
          "analyzer": {
            "char_filters": [
              {
                "name": "ascii_folding",
                "options": null
              },
              {
                "name": "unicode_normalize",
                "options": {
                  "form": "NFKC"
                }
              }
            ],
            "tokenizer": {
              "name": "single_token",
              "options": null
            },
            "token_filters": [
              {
                "name": "lower_case",
                "options": null
              }
            ]
          }
        },
        "redirect": {
          "type": "text",
          "options": {
            "index": true,
            "store": true,
            "term_positions": false,
            "highlight": false,
            "sortable": false,
            "aggregatable": true
          },
          "analyzer": {
            "char_filters": [
              {
                "name": "ascii_folding",
                "options": null
              },
              {
                "name": "unicode_normalize",
                "options": {
                  "form": "NFKC"
                }
              }
            ],
            "tokenizer": {
              "name": "single_token",
              "options": null
            },
            "token_filters": [
              {
                "name": "lower_case",
                "options": null
              }
            ]
          }
        },
        "text": {
          "type": "text",
          "options": {
            "index": true,
            "store": true,
            "term_positions": true,
            "highlight": true,
            "sortable": false,
            "aggregatable": false
          },
          "analyzer": {
            "char_filters": [
              {
                "name": "ascii_folding",
                "options": null
              },
              {
                "name": "unicode_normalize",
                "options": {
                  "form": "NFKC"
                }
              }
            ],
            "tokenizer": {
              "name": "unicode",
              "options": null
            },
            "token_filters": [
              {
                "name": "lower_case",
                "options": null
              }
            ]
          }
        },
        "timestamp": {
          "type": "datetime",
          "options": {
            "index": true,
            "store": true,
            "term_positions": false,
            "highlight": false,
            "sortable": true,
            "aggregatable": true
          },
          "analyzer": {
            "char_filters": null,
            "tokenizer": {
              "name": "",
              "options": null
            },
            "token_filters": null
          }
        },
        "title": {
          "type": "text",
          "options": {
            "index": true,
            "store": true,
            "term_positions": true,
            "highlight": true,
            "sortable": false,
            "aggregatable": false
          },
          "analyzer": {
            "char_filters": [
              {
                "name": "ascii_folding",
                "options": null
              },
              {
                "name": "unicode_normalize",
                "options": {
                  "form": "NFKC"
                }
              }
            ],
            "tokenizer": {
              "name": "unicode",
              "options": null
            },
            "token_filters": [
              {
                "name": "lower_case",
                "options": null
              }
            ]
          }
        },
        "url": {
          "type": "text",
          "options": {
            "index": true,
            "store": true,
            "term_positions": false,
            "highlight": false,
            "sortable": false,
            "aggregatable": true
          },
          "analyzer": {
            "char_filters": [
              {
                "name": "ascii_folding",
                "options": null
              },
              {
                "name": "unicode_normalize",
                "options": {
                  "form": "NFKC"
                }
              }
            ],
            "tokenizer": {
              "name": "single_token",
              "options": null
            },
            "token_filters": [
              {
                "name": "lower_case",
                "options": null
              }
            ]
          }
        }
      },
      "index_uri": "minio://phalanx-indexes/example",
      "shards": {
        "shard-0kZFXhrZ": {
          "shard_lock_uri": "etcd://phalanx-locks/example/shard-0kZFXhrZ",
          "shard_uri": "minio://phalanx-indexes/example/shard-0kZFXhrZ"
        },
        "shard-FretswBz": {
          "shard_lock_uri": "etcd://phalanx-locks/example/shard-FretswBz",
          "shard_uri": "minio://phalanx-indexes/example/shard-FretswBz"
        },
        "shard-Lf7rlqwd": {
          "shard_lock_uri": "etcd://phalanx-locks/example/shard-Lf7rlqwd",
          "shard_uri": "minio://phalanx-indexes/example/shard-Lf7rlqwd"
        },
        "shard-WvUd7lWm": {
          "shard_lock_uri": "etcd://phalanx-locks/example/shard-WvUd7lWm",
          "shard_uri": "minio://phalanx-indexes/example/shard-WvUd7lWm"
        },
        "shard-f4f6jbCi": {
          "shard_lock_uri": "etcd://phalanx-locks/example/shard-f4f6jbCi",
          "shard_uri": "minio://phalanx-indexes/example/shard-f4f6jbCi"
        },
        "shard-kCPFctnO": {
          "shard_lock_uri": "etcd://phalanx-locks/example/shard-kCPFctnO",
          "shard_uri": "minio://phalanx-indexes/example/shard-kCPFctnO"
        }
      }
    }
  },
  "nodes": {
    "node-09upDjwO": {
      "addr": "0.0.0.0",
      "meta": {
        "grpc_port": 5001,
        "http_port": 8001,
        "roles": [
          "indexer",
          "searcher"
        ]
      },
      "port": 2001,
      "state": "alive"
    },
    "node-8giCE4QA": {
      "addr": "0.0.0.0",
      "meta": {
        "grpc_port": 5002,
        "http_port": 8002,
        "roles": [
          "indexer",
          "searcher"
        ]
      },
      "port": 2002,
      "state": "alive"
    },
    "node-Rh6kmVO8": {
      "addr": "0.0.0.0",
      "meta": {
        "grpc_port": 5000,
        "http_port": 8000,
        "roles": [
          "indexer",
          "searcher"
        ]
      },
      "port": 2000,
      "state": "alive"
    }
  },
  "searcher_assignment": {
    "example": {
      "shard-0kZFXhrZ": [
        "node-8giCE4QA",
        "node-Rh6kmVO8",
        "node-09upDjwO"
      ],
      "shard-FretswBz": [
        "node-Rh6kmVO8",
        "node-09upDjwO",
        "node-8giCE4QA"
      ],
      "shard-Lf7rlqwd": [
        "node-Rh6kmVO8",
        "node-09upDjwO",
        "node-8giCE4QA"
      ],
      "shard-WvUd7lWm": [
        "node-Rh6kmVO8",
        "node-8giCE4QA",
        "node-09upDjwO"
      ],
      "shard-f4f6jbCi": [
        "node-Rh6kmVO8",
        "node-8giCE4QA",
        "node-09upDjwO"
      ],
      "shard-kCPFctnO": [
        "node-8giCE4QA",
        "node-Rh6kmVO8",
        "node-09upDjwO"
      ]
    }
  }
}
```


## Add / Update documents

```
% cat ./testdata/enwiki-20211201-1000.jsonl | jq -c -r '. |  ("https://en.wikipedia.org/wiki/" + .title | gsub(" "; "_")) as $url | [.links[].PageName] as $links | [.media[].PageName] as $media | . |= .+ {"links": $links, "media": $media, "url": $url}' | ./bin/phalanx_docs.sh --id-field=id | curl -XPUT -H 'Content-type: application/x-ndjson' http://localhost:8000/v1/indexes/enwiki/documents --data-binary @-
```


## Search

```
% curl -XPOST -H 'Content-type: application/json' http://localhost:8000/v1/indexes/enwiki/_search --data-binary '
{
    "query": {
        "type": "boolean",
        "options": {
            "must": [
                {
                    "type": "query_string",
                    "options": {
                        "query": "search engine"
                    }
                }
            ],
            "min_should": 1,
            "boost": 1.0
        }
    },
    "start": 0,
    "num": 10,
    "sort_by": "-_score",
    "fields": [
        "title",
        "text",
        "url"
    ],
    "aggregations": {
        "timestamp_date_range": {
            "type": "date_range",
            "options": {
                "field": "_timestamp",
                "ranges": {
                    "year_before_last": {
                        "start": "2020-01-01T00:00:00Z",
                        "end": "2021-01-01T00:00:00Z"
                    },
                    "last_year": {
                        "start": "2021-01-01T00:00:00Z",
                        "end": "2022-01-01T00:00:00Z"
                    },
                    "this_year": {
                        "start": "2022-01-01T00:00:00Z",
                        "end": "2023-01-01T00:00:00Z"
                    }
                }
            }
        }
    }
}
' | jq .
```

```json
{
  "aggregations": {
    "timestamp_date_range": {
      "last_year": 0,
      "this_year": 16,
      "year_before_last": 0
    }
  },
  "documents": [
    {
      "fields": {
        "text": [
          "The Analytical Engine was a proposed mechanical general-purpose computer designed by English mathematician and computer pioneer Charles Babbage. It was first described in 1837 as the successor to Babbage's difference engine, which was a design for a simpler mechanical computer.\nThe Analytical Engine incorporated an arithmetic logic unit, control flow in the form of conditional branching and loops, and integrated memory, making it the first design for a general-purpose computer that could be described in modern terms as Turing-complete. In other words, the logical structure of the Analytical Engine was essentially the same as that which has dominated computer design in the electronic era. It was not until 1941 that Konrad Zuse built the first general-purpose computer, Z3, more than a century after Babbage had proposed the pioneering Analytical Engine in 1837.\nDuring this project, Babbage realised that a much more general design, the Analytical Engine, was possible. The work on the design of the Analytical Engine started in c. 1833.\nThe input, consisting of programs (\"formulae\") and data, was to be provided to the machine via punched cards, a method being used at the time to direct mechanical looms such as the Jacquard loom. For output, the machine would have a printer, a curve plotter, and a bell. The machine would also be able to punch numbers onto cards to be read in later. It employed ordinary base-10 fixed-point arithmetic.\nThere was to be a store (that is, a memory) capable of holding 1,000 numbers of 40 decimal digits each (ca. 16.6 kB). An arithmetic unit (the \"mill\") would be able to perform all four arithmetic operations, plus comparisons and optionally square roots. Initially (1838) it was conceived as a difference engine curved back upon itself, in a generally circular layout, with the long store exiting off to one side. Later drawings (1858) depict a regularised grid layout. Like the central processing unit (CPU) in a modern computer, the mill would rely upon its own internal procedures, to be stored in the form of pegs inserted into rotating drums called \"barrels\", to carry out some of the more complex instructions the user's program might specify.\nThe programming language to be employed by users was akin to modern day assembly languages. Loops and conditional branching were possible, and so the language as conceived would have been Turing-complete as later defined by Alan Turing. Three different types of punch cards were used: one for arithmetical operations, one for numerical constants, and one for load and store operations, transferring numbers from the store to the arithmetical unit or back. There were three separate readers for the three types of cards. Babbage developed some two dozen programs for the Analytical Engine between 1837 and 1840, and one program later. These programs treat polynomials, iterative formulas, Gaussian elimination, and Bernoulli numbers.\nIn 1842, the Italian mathematician Luigi Federico Menabrea published a description of the engine in French, based on lectures Babbage gave when he visited Turin in 1840. In 1843, the description was translated into English and extensively annotated by Ada Lovelace, who had become interested in the engine eight years earlier. In recognition of her additions to Menabrea's paper, which included a way to calculate Bernoulli numbers using the machine (widely considered to be the first complete computer program), she has been described as the first computer programmer."
        ],
        "title": [
          "Analytical Engine"
        ],
        "url": [
          "https://en.wikipedia.org/wiki/Analytical_Engine"
        ]
      },
      "id": "1271",
      "score": 4.294355129514904,
      "timestamp": 1644822812843749600
    },
    {
      "fields": {
        "text": [
          " Ada Lovelace#Ada Byron's notes on the analytical engine"
        ],
        "title": [
          "Ada Byron's notes on the analytical engine"
        ],
        "url": [
          "https://en.wikipedia.org/wiki/Ada_Byron's_notes_on_the_analytical_engine"
        ]
      },
      "id": "1311",
      "score": 4.215404371132957,
      "timestamp": 1644822812919017000
    },
    {
      "fields": {
        "text": [
          "Augusta Ada King, Countess of Lovelace (née Byron; 10 December 1815 – 27 November 1852) was an English mathematician and writer, chiefly known for her work on Charles Babbage's proposed mechanical general-purpose computer, the Analytical Engine. She was the first to recognise that the machine had applications beyond pure calculation, and to have published the first algorithm intended to be carried out by such a machine. As a result, she is often regarded as the first computer programmer.\nAda Byron was the only child of poet Lord Byron and mathematician Lady Byron. All of Byron's other children were born out of wedlock to other women. He died in Greece when Ada was eight years old. Her mother remained bitter and promoted Ada's interest in mathematics and logic in an effort to prevent her from developing her father's perceived insanity. Despite this, Ada remained interested in him, naming her two sons Byron and Gordon. Upon her death, she was buried next to him at her request. Although often ill in her childhood, Ada pursued her studies assiduously. She married William King in 1835. King was made Earl of Lovelace in 1838, Ada thereby becoming Countess of Lovelace.\nHer educational and social exploits brought her into contact with scientists such as Andrew Crosse, Charles Babbage, Sir David Brewster, Charles Wheatstone, Michael Faraday and the author Charles Dickens, contacts which she used to further her education. Ada described her approach as \"poetical science\" and herself as an \"Analyst (& Metaphysician)\".\nWhen she was a teenager (18), her mathematical talents led her to a long working relationship and friendship with fellow British mathematician Charles Babbage, who is known as \"the father of computers\". She was in particular interested in Babbage's work on the Analytical Engine. Lovelace first met him in June 1833, through their mutual friend, and her private tutor, Mary Somerville.\nBetween 1842 and 1843, Ada translated an article by Italian military engineer Luigi Menabrea about the Analytical Engine, supplementing it with an elaborate set of notes, simply called \"Notes\". Lovelace's notes are important in the early history of computers, containing what many consider to be the first computer program—that is, an algorithm designed to be carried out by a machine. Other historians reject this perspective and point out that Babbage's personal notes from the years 1836/1837 contain the first programs for the engine. She also developed a vision of the capability of computers to go beyond mere calculating or number-crunching, while many others, including Babbage himself, focused only on those capabilities. Her mindset of \"poetical science\" led her to ask questions about the Analytical Engine (as shown in her notes) examining how individuals and society relate to technology as a collaborative tool.\nShe died of uterine cancer in 1852 at the age of 36."
        ],
        "title": [
          "Ada Lovelace"
        ],
        "url": [
          "https://en.wikipedia.org/wiki/Ada_Lovelace"
        ]
      },
      "id": "974",
      "score": 3.4604735722185667,
      "timestamp": 1644822812408895000
    },
    {
      "fields": {
        "text": [
          "Adder may refer to:\n Any of several groups of venomous snakes\n Vipera berus, the common European adder, found in Europe and northern Asia\n Adder (electronics), an electronic circuit designed to do addition\n AA-12 Adder, the NATO name for the R-77, a Russian air-to-air missile\n HMS Adder, any of seven ships of the Royal Navy\n USS Adder (SS-3), an early US submarine\n Adder Technology, a manufacturing company\n Addition, a mathematical operation\n Armstrong Siddeley Adder, a late 1940s British turbojet engine\n Blackadder, a series of BBC sitcoms\n Golden Axe: The Revenge of Death Adder, a video game"
        ],
        "title": [
          "Adder"
        ],
        "url": [
          "https://en.wikipedia.org/wiki/Adder"
        ]
      },
      "id": "1538",
      "score": 2.9936382329174775,
      "timestamp": 1644822813067768800
    },
    {
      "fields": {
        "text": [
          "The Royal Antigua and Barbuda Defence Force is the armed forces of Antigua and Barbuda. The RABDF has responsibility for several different roles: internal security, prevention of drug smuggling, the protection and support of fishing rights, prevention of marine pollution, search and rescue, ceremonial duties, assistance to government programs, provision of relief during natural disasters, assistance in the maintenance of essential services, and support of the police in maintaining law and order.\nThe RABDF is one of the world's smallest militaries, consisting of 245 personnel. It is much better equipped for fulfilling its civil roles as opposed to providing a deterrence against would-be aggressors or in defending the nation during a war."
        ],
        "title": [
          "Royal Antigua and Barbuda Defence Force"
        ],
        "url": [
          "https://en.wikipedia.org/wiki/Royal_Antigua_and_Barbuda_Defence_Force"
        ]
      },
      "id": "1074",
      "score": 2.8903576929024535,
      "timestamp": 1644822812495888400
    },
    {
      "fields": {
        "text": [
          "In computing, an applet is any small application that performs one specific task that runs within the scope of a dedicated widget engine or a larger program, often as a plug-in. The term is frequently used to refer to a Java applet, a program written in the Java programming language that is designed to be placed on a web page. Applets are typical examples of transient and auxiliary applications that don't monopolize the user's attention. Applets are not full-featured application programs, and are intended to be easily accessible."
        ],
        "title": [
          "Applet"
        ],
        "url": [
          "https://en.wikipedia.org/wiki/Applet"
        ]
      },
      "id": "1202",
      "score": 2.8648928474658204,
      "timestamp": 1644822812706504200
    },
    {
      "fields": {
        "text": [
          " (; born , ; 16 April 1844 – 12 October 1924) was a French poet, journalist, and novelist with several best-sellers. Ironic and skeptical, he was considered in his day the ideal French man of letters. He was a member of the Académie française, and won the 1921 Nobel Prize in Literature \"in recognition of his brilliant literary achievements, characterized as they are by a nobility of style, a profound human sympathy, grace, and a true Gallic temperament\".\nFrance is also widely believed to be the model for narrator Marcel's literary idol Bergotte in Marcel Proust's In Search of Lost Time."
        ],
        "title": [
          "Anatole France"
        ],
        "url": [
          "https://en.wikipedia.org/wiki/Anatole_France"
        ]
      },
      "id": "1057",
      "score": 2.540780345712897,
      "timestamp": 1644822812700458800
    },
    {
      "fields": {
        "text": [
          "Argo Navis (the Ship Argo), or simply Argo, was a large constellation in the southern sky. The genitive was \"Argus Navis\", abbreviated \"Arg\". Flamsteed and other early modern astronomers called it Navis (the Ship), genitive \"Navis\", abbreviated \"Nav\".\nThe constellation proved to be of unwieldy size, as it was 28% larger than the next largest constellation and had more than 160 easily visible stars. The 1755 catalogue of Nicolas Louis de Lacaille divided it into the three modern constellations that occupy much of the same area: Carina (the keel), Puppis (the poop deck) and Vela (the sails).\nArgo derived from the ship Argo in Greek mythology, sailed by Jason and the Argonauts to Colchis in search of the Golden Fleece. Due to precession of the equinoxes, the position of the stars from Earth's viewpoint has shifted southward, and though most of the constellation was visible in Classical times, the constellation is now not easily visible from most of the northern hemisphere. All the stars of Argo Navis are easily visible from the tropics southward, and pass near zenith from southern temperate latitudes. The brightest of these is Canopus (α Carinae), the second-brightest night-time star, now assigned to Carina."
        ],
        "title": [
          "Argo Navis"
        ],
        "url": [
          "https://en.wikipedia.org/wiki/Argo_Navis"
        ]
      },
      "id": "1924",
      "score": 2.331200750903565,
      "timestamp": 1644822813221983500
    },
    {
      "fields": {
        "text": [
          "An assembly line is a manufacturing process (often called a progressive assembly) in which parts (usually interchangeable parts) are added as the semi-finished assembly moves from workstation to workstation where the parts are added in sequence until the final assembly is produced. By mechanically moving the parts to the assembly work and moving the semi-finished assembly from work station to work station, a finished product can be assembled faster and with less labor than by having workers carry parts to a stationary piece for assembly.\nAssembly lines are common methods of assembling complex items such as automobiles and other transportation equipment, household appliances and electronic goods.\nWorkers in charge of the works of assembly line are called assemblers."
        ],
        "title": [
          "Assembly line"
        ],
        "url": [
          "https://en.wikipedia.org/wiki/Assembly_line"
        ]
      },
      "id": "1146",
      "score": 2.3034866993609384,
      "timestamp": 1644822812771819800
    },
    {
      "fields": {
        "text": [
          "André Paul Guillaume Gide (; 22 November 1869 – 19 February 1951) was a French author and winner of the Nobel Prize in Literature (in 1947). Gide's career ranged from its beginnings in the symbolist movement, to the advent of anticolonialism between the two World Wars. The author of more than fifty books, at the time of his death his obituary in The New York Times described him as \"France's greatest contemporary man of letters\" and \"judged the greatest French writer of this century by the literary cognoscenti.\"\nKnown for his fiction as well as his autobiographical works, Gide exposed to public view the conflict and eventual reconciliation of the two sides of his personality (characterized by a Protestant austerity and a transgressive sexual adventurousness, respectively), which a strict and moralistic education had helped set at odds. Gide's work can be seen as an investigation of freedom and empowerment in the face of moralistic and puritanical constraints, and centers on his continuous effort to achieve intellectual honesty. His self-exploratory texts reflect his search of how to be fully oneself, including owning one's sexual nature, without at the same time betraying one's values. His political activity was shaped by the same ethos, as indicated by his repudiation of communism after his 1936 journey to the USSR."
        ],
        "title": [
          "André Gide"
        ],
        "url": [
          "https://en.wikipedia.org/wiki/André_Gide"
        ]
      },
      "id": "1058",
      "score": 2.0794730555180783,
      "timestamp": 1644822812633650400
    }
  ],
  "hits": 16,
  "index_name": "enwiki"
}
```


## Delete index

```
% curl -XDELETE http://localhost:8000/v1/indexes/enwiki
```
