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
        "id": {
          "type": "numeric",
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
        "text": {
          "type": "text",
          "options": {
            "index": true,
            "store": true,
            "term_positions": true,
            "highlight": true,
            "sortable": true,
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
% cat ./testdata/enwiki-20211201-1000.jsonl | jq -c -r '. | [.links[].PageName] as $links | [.media[].PageName] as $media | . |= .+ {"links": $links, "media": $media}' | ./bin/phalanx_docs.sh --id-field=id | curl -XPUT -H 'Content-type: application/x-ndjson' http://localhost:8000/v1/indexes/enwiki/documents --data-binary @-
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
                        "query": "search"
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
        "*"
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
      "this_year": 6,
      "year_before_last": 0
    }
  },
  "documents": [
    {
      "fields": {
        "categories": "Use_dmy_dates_from_June_2013",
        "external_links": "http://www.uwichill.edu.bb/bnccde/antigua/conference/papers/phillips.html",
        "id": 1074,
        "links": "Military units and formations established in 1981",
        "media": "A Royal Antigua and Barbuda Defense Force Coast Guard 380X Defender All-Weather Interceptor high-speed boat breeches the wake of another vessel May 24, 2013, during the maritime operations phase of Tradewinds 130524-N-HP195-034.jpg",
        "redirect": "",
        "text": "\n\n\n\n\n\nThe Royal Antigua and Barbuda Defence Force is the armed forces of Antigua and Barbuda. The RABDF has responsibility for several different roles: internal security, prevention of drug smuggling, the protection and support of fishing rights, prevention of marine pollution, search and rescue, ceremonial duties, assistance to government programs, provision of relief during natural disasters, assistance in the maintenance of essential services, and support of the police in maintaining law and order.\n\n\nThe RABDF is one of the world's smallest militaries, consisting of 245 personnel. It is much better equipped for fulfilling its civil roles as opposed to providing a deterrence against would-be aggressors or in defending the nation during a war.\n\n\n\n",
        "timestamp": "2021-10-22T10:59:08Z",
        "title": "Royal Antigua and Barbuda Defence Force"
      },
      "id": "1074",
      "score": 2.8849718777597833,
      "timestamp": 1644763316585247000
    },
    {
      "fields": {
        "categories": "Writers_from_Paris",
        "external_links": "http://www.litteratureaudio.com/livres-audio-gratuits-mp3/tag/anatole-france/",
        "id": 1057,
        "links": "Christian novelists",
        "media": "Speaker Icon.svg",
        "redirect": "",
        "text": "\n\n\n\n\n\n\n\n (; born , ; 16 April 1844 – 12 October 1924) was a French poet, journalist, and novelist with several best-sellers. Ironic and skeptical, he was considered in his day the ideal French man of letters. He was a member of the Académie française, and won the 1921 Nobel Prize in Literature \"in recognition of his brilliant literary achievements, characterized as they are by a nobility of style, a profound human sympathy, grace, and a true Gallic temperament\".\n\n\nFrance is also widely believed to be the model for narrator Marcel's literary idol Bergotte in Marcel Proust's In Search of Lost Time.\n\n\n\n",
        "timestamp": "2021-11-01T05:10:06Z",
        "title": "Anatole France"
      },
      "id": "1057",
      "score": 2.552997403589038,
      "timestamp": 1644763316588002000
    },
    {
      "fields": {
        "categories": "Short_description_is_different_from_Wikidata",
        "external_links": "https://iconographic.warburg.sas.ac.uk/vpc/VPC_search/subcats.php?cat_1=9&cat_2=71&cat_3=32&cat_4=1317&cat_5=991",
        "id": 1924,
        "links": "Former constellations",
        "media": "Image:Argo Navis Hevelius.jpg",
        "redirect": "",
        "text": "\n\nArgo Navis (the Ship Argo), or simply Argo, was a large constellation in the southern sky. The genitive was \"Argus Navis\", abbreviated \"Arg\". Flamsteed and other early modern astronomers called it Navis (the Ship), genitive \"Navis\", abbreviated \"Nav\".\n\n\nThe constellation proved to be of unwieldy size, as it was 28% larger than the next largest constellation and had more than 160 easily visible stars. The 1755 catalogue of Nicolas Louis de Lacaille divided it into the three modern constellations that occupy much of the same area: Carina (the keel), Puppis (the poop deck) and Vela (the sails).\n\n\nArgo derived from the ship Argo in Greek mythology, sailed by Jason and the Argonauts to Colchis in search of the Golden Fleece. Due to precession of the equinoxes, the position of the stars from Earth's viewpoint has shifted southward, and though most of the constellation was visible in Classical times, the constellation is now not easily visible from most of the northern hemisphere. All the stars of Argo Navis are easily visible from the tropics southward, and pass near zenith from southern temperate latitudes. The brightest of these is Canopus (α Carinae), the second-brightest night-time star, now assigned to Carina.\n\n\n\n",
        "timestamp": "2021-10-12T08:41:22Z",
        "title": "Argo Navis"
      },
      "id": "1924",
      "score": 2.4393900530430184,
      "timestamp": 1644763316613395500
    },
    {
      "fields": {
        "categories": "Writers_from_Paris",
        "external_links": "https://web.archive.org/web/20071006001903/http://www.societe-jersiaise.org/whitsco/mader1.htm",
        "id": 1058,
        "links": "20th-century LGBT people",
        "media": "Image:Gide by Laurens.jpg",
        "redirect": "",
        "text": "\n\n\n\n\n\nAndré Paul Guillaume Gide (; 22 November 1869 – 19 February 1951) was a French author and winner of the Nobel Prize in Literature (in 1947). Gide's career ranged from its beginnings in the symbolist movement, to the advent of anticolonialism between the two World Wars. The author of more than fifty books, at the time of his death his obituary in The New York Times described him as \"France's greatest contemporary man of letters\" and \"judged the greatest French writer of this century by the literary cognoscenti.\"\n\n\nKnown for his fiction as well as his autobiographical works, Gide exposed to public view the conflict and eventual reconciliation of the two sides of his personality (characterized by a Protestant austerity and a transgressive sexual adventurousness, respectively), which a strict and moralistic education had helped set at odds. Gide's work can be seen as an investigation of freedom and empowerment in the face of moralistic and puritanical constraints, and centers on his continuous effort to achieve intellectual honesty. His self-exploratory texts reflect his search of how to be fully oneself, including owning one's sexual nature, without at the same time betraying one's values. His political activity was shaped by the same ethos, as indicated by his repudiation of communism after his 1936 journey to the USSR.\n\n\n\n",
        "timestamp": "2021-11-22T11:09:31Z",
        "title": "André Gide"
      },
      "id": "1058",
      "score": 2.167467704764721,
      "timestamp": 1644763316587237000
    },
    {
      "fields": {
        "categories": "Wikipedia_indefinitely_semi-protected_pages",
        "external_links": "https://archive.org/details/encyclopediaofas0000embr",
        "id": 689,
        "links": "Continents",
        "media": "20091002 Hong Kong 6269.jpg",
        "redirect": "",
        "text": "\n\n\n\n\n\n\n\n\n\n\nAsia () is Earth's largest and most populous continent, located primarily in the Eastern and Northern Hemispheres. It shares the continental landmass of Eurasia with the continent of Europe and the continental landmass of Afro-Eurasia with both Europe and Africa. Asia covers an area of , about 30% of Earth's total land area and 8.7% of the Earth's total surface area. The continent, which has long been home to the majority of the human population, was the site of many of the first civilizations. Its 4.5 billion people () constitute roughly 60% of the world's population.\n\n\nIn general terms, Asia is bounded on the east by the Pacific Ocean, on the south by the Indian Ocean, and on the north by the Arctic Ocean. The border of Asia with Europe is a historical and cultural construct, as there is no clear physical and geographical separation between them. It is somewhat arbitrary and has moved since its first conception in classical antiquity. The division of Eurasia into two continents reflects East–West cultural, linguistic, and ethnic differences, some of which vary on a spectrum rather than with a sharp dividing line. The most commonly accepted boundaries place Asia to the east of the Suez Canal separating it from Africa; and to the east of the Turkish Straits, the Ural Mountains and Ural River, and to the south of the Caucasus Mountains and the Caspian and Black Seas, separating it from Europe.\n\n\nChina and India alternated in being the largest economies in the world from 1 to 1800 CE. China was a major economic power and attracted many to the east, and for many the legendary wealth and prosperity of the ancient culture of India personified Asia, attracting European commerce, exploration and colonialism. The accidental discovery of a trans-Atlantic route from Europe to America by Columbus while in search for a route to India demonstrates this deep fascination. The Silk Road became the main east–west trading route in the Asian hinterlands while the Straits of Malacca stood as a major sea route. Asia has exhibited economic dynamism (particularly East Asia) as well as robust population growth during the 20th century, but overall population growth has since fallen. Asia was the birthplace of most of the world's mainstream religions including Hinduism, Zoroastrianism, Judaism, Jainism, Buddhism, Confucianism, Taoism, Christianity, Islam, Sikhism, as well as many other religions.\n\n\nGiven its size and diversity, the concept of Asia—a name dating back to classical antiquity—may actually have more to do with human geography than physical geography. Asia varies greatly across and within its regions with regard to ethnic groups, cultures, environments, economics, historical ties and government systems. It also has a mix of many different climates ranging from the equatorial south via the hot desert in the Middle East, temperate areas in the east and the continental centre to vast subarctic and polar areas in Siberia.\n\n\n\n",
        "timestamp": "2021-11-12T18:57:25Z",
        "title": "Asia"
      },
      "id": "689",
      "score": 1.1987013812725258,
      "timestamp": 1644763316573834200
    },
    {
      "fields": {
        "categories": "Wikipedia_indefinitely_semi-protected_pages",
        "external_links": "https://www.bbc.co.uk/programmes/p003k9fc",
        "id": 1164,
        "links": "Computational fields of study",
        "media": "Capek play.jpg",
        "redirect": "",
        "text": "\n\n\n\n\n\n\nArtificial intelligence (AI) is intelligence demonstrated by machines, as opposed to natural intelligence displayed by animals including humans.\nLeading AI textbooks define the field as the study of \"intelligent agents\": any system that perceives its environment and takes actions that maximize its chance of achieving its goals.\nSome popular accounts use the term \"artificial intelligence\" to describe machines that mimic \"cognitive\" functions that humans associate with the human mind, such as \"learning\" and \"problem solving\", however, this definition is rejected by major AI researchers.\n\n\n\n\nAI applications include advanced web search engines (e.g., Google), recommendation systems (used by YouTube, Amazon and Netflix), understanding human speech (such as Siri and Alexa), self-driving cars (e.g., Tesla), automated decision-making and competing at the highest level in strategic game systems (such as chess and Go).\nAs machines become increasingly capable, tasks considered to require \"intelligence\" are often removed from the definition of AI, a phenomenon known as the AI effect. For instance, optical character recognition is frequently excluded from things considered to be AI, having become a routine technology.\n\n\n\n\nArtificial intelligence was founded as an academic discipline in 1956, and in the years since has experienced several waves of optimism,\nand have been common in fiction, as in Mary Shelley's Frankenstein or Karel Čapek's R.U.R. These characters and their fates raised many of the same issues now discussed in the ethics of artificial intelligence.\n\n\n\n\nThe study of mechanical or \"formal\" reasoning began with philosophers and mathematicians in antiquity. The study of mathematical logic led directly to Alan Turing's theory of computation, which suggested that a machine, by shuffling symbols as simple as \"0\" and \"1\", could simulate any conceivable act of mathematical deduction. This insight that digital computers can simulate any process of formal reasoning is known as the Church–Turing thesis.\n\n\n\n\nThe Church-Turing thesis, along with concurrent discoveries in neurobiology, information theory and cybernetics, led researchers to consider the possibility of building an electronic brain.\nThe first work that is now generally recognized as AI was McCullouch and Pitts' 1943 formal design for Turing-complete \"artificial neurons\".\n\n\n\n\nWhen access to digital computers became possible in the mid-1950s, AI research began to explore the possibility that human intelligence could be reduced to step-by-step symbol manipulation, known as Symbolic AI or GOFAI. Approaches based on cybernetics or artificial neural networks were abandoned or pushed into the background.\n\n\n\n\nThe field of AI research was born at a workshop at Dartmouth College in 1956.\nThe attendees became the founders and leaders of AI research.\nThey and their students produced programs that the press described as \"astonishing\":\ncomputers were learning checkers strategies, solving word problems in algebra, proving logical theorems and speaking English.\nBy the middle of the 1960s, research in the U.S. was heavily funded by the Department of Defense\nand laboratories had been established around the world.\n\n\n\n\nResearchers in the 1960s and the 1970s were convinced that symbolic approaches would eventually succeed in creating a machine with artificial general intelligence and considered this the goal of their field.\nHerbert Simon predicted, \"machines will be capable, within twenty years, of doing any work a man can do\".\nMarvin Minsky agreed, writing, \"within a generation ... the problem of creating 'artificial intelligence' will substantially be solved\".\n\n\n\n\nThey failed to recognize the difficulty of some of the remaining tasks. Progress slowed and in 1974, in response to the criticism of Sir James Lighthill\nand ongoing pressure from the US Congress to fund more productive projects, both the U.S. and British governments cut off exploratory research in AI. The next few years would later be called an \"AI winter\", a period when obtaining funding for AI projects was difficult.\n\n\n\n\n\n\nIn the early 1980s, AI research was revived by the commercial success of expert systems,\na form of AI program that simulated the knowledge and analytical skills of human experts. By 1985, the market for AI had reached over a billion dollars. At the same time, Japan's fifth generation computer project inspired the U.S and British governments to restore funding for academic research.\nHowever, beginning with the collapse of the Lisp Machine market in 1987, AI once again fell into disrepute, and a second, longer-lasting winter began.\n\n\n\n\nMany researchers began to doubt that the symbolic approach would be able to imitate all the processes of human cognition, especially perception, robotics, learning and pattern recognition. A number of researchers began to look into \"sub-symbolic\" approaches to specific AI problems. Robotics researchers, such as Rodney Brooks, rejected symbolic AI and focused on the basic engineering problems that would allow robots to move, survive, and learn their environment.\nInterest in neural networks and \"connectionism\" was revived by Geoffrey Hinton, David Rumelhart and others in the middle of the 1980s.\nSoft computing tools were developed in the 80s, such as neural networks, fuzzy systems, Grey system theory, evolutionary computation and many tools drawn from statistics or mathematical optimization.\n\n\n\n\nAI gradually restored its reputation in the late 1990s and early 21st century by finding specific solutions to specific problems. The narrow focus allowed researchers to produce verifiable results, exploit more mathematical methods, and collaborate with other fields (such as statistics, economics and mathematics).\nBy 2000, solutions developed by AI researchers were being widely used, although in the 1990s they were rarely described as \"artificial intelligence\".\n\n\n\n\nFaster computers, algorithmic improvements, and access to large amounts of data enabled advances in machine learning and perception; data-hungry deep learning methods started to dominate accuracy benchmarks around 2012.\nAccording to Bloomberg's Jack Clark, 2015 was a landmark year for artificial intelligence, with the number of software projects that use AI within Google increased from a \"sporadic usage\" in 2012 to more than 2,700 projects. He attributes this to an increase in affordable neural networks, due to a rise in cloud computing infrastructure and to an increase in research tools and datasets. In a 2017 survey, one in five companies reported they had \"incorporated AI in some offerings or processes\". The amount of research into AI (measured by total publications) increased by 50% in the years 2015–2019.\n\n\n\n\nNumerous academic researchers became concerned that AI was no longer pursuing the original goal of creating versatile, fully intelligent machines. Much of current research involves statistical AI, which is overwhelmingly used to solve specific problems, even highly successful techniques such as deep learning. This concern has led to the subfield artificial general intelligence (or \"AGI\"), which had several well-funded institutions by the 2010s.\n\n\n\n",
        "timestamp": "2021-11-25T04:55:17Z",
        "title": "Artificial intelligence"
      },
      "id": "1164",
      "score": 0.8702074066927157,
      "timestamp": 1644763316586224400
    }
  ],
  "hits": 6,
  "index_name": "enwiki"
}
```


## Delete index

```
% curl -XDELETE http://localhost:8000/v1/indexes/enwiki
```
