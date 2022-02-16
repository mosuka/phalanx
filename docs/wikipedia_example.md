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

```
% ./bin/phalanx --index-metastore-uri=etcd://phalanx-metadata
```


## Create index with MinIO and etcd

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


## Add / Update documents

```
% cat ./testdata/enwiki-20211201-1000.jsonl | jq -c -r '. |  ("https://en.wikipedia.org/wiki/" + .title | gsub(" "; "_")) as $url | [.links[].PageName] as $links | [.media[].PageName] as $media | . |= .+ {"links": $links, "media": $media, "url": $url}' | ./bin/phalanx_docs.sh --id-field=id | curl -XPUT -H 'Content-type: application/x-ndjson' http://localhost:8000/v1/indexes/enwiki/documents --data-binary @-
```


## Search

```json
% curl -XPOST -H 'Content-type: application/json' http://localhost:8000/v1/indexes/enwiki/_search --data-binary '
{
    "query": {
        "type": "boolean",
        "options": {
            "must": [
                {
                    "type": "query_string",
                    "options": {
                        "query": "+text:search"
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
    },
    "highlights": {
        "title": {
            "highlighter": {
                "type": "html",
                "options": {
                    "fragment_size": 100,
                    "pre_tag": "<mark>",
                    "post_tag": "</mark>"
                }
            },
            "num": 1
        },
        "text": {
            "highlighter": {
              "type": "html",
              "options": {
                  "fragment_size": 200,
                  "pre_tag": "<mark>",
                  "post_tag": "</mark>"
              }
            },
            "num": 3
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
      "highlights": {
        "text": [
          "…f style, a profound human sympathy, grace, and a true Gallic temperament&#34;.\nFrance is also widely believed to be the model for narrator Marcel&#39;s literary idol Bergotte in Marcel Proust&#39;s<mark> In Se</mark>arch of L…"
        ],
        "title": [
          "Anatole France"
        ]
      },
      "id": "1057",
      "score": 2.5689701938929193,
      "timestamp": 1644936891227494000
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
      "highlights": {
        "text": [
          "…of drug smuggling, the protection and support of fishing rights, prevention of marine pollution, <mark>search</mark> and rescue, ceremonial duties, assistance to government programs, provision of relief during nat…"
        ],
        "title": [
          "Royal Antigua and Barbuda Defence Force"
        ]
      },
      "id": "1074",
      "score": 2.4424078634334503,
      "timestamp": 1644936891224384500
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
      "highlights": {
        "text": [
          "…rs on his continuous effort to achieve intellectual honesty. His self-exploratory texts reflect h<mark>is sea</mark>rch of how to be fully oneself, including owning one&#39;s sexual nature, without at the same time be…"
        ],
        "title": [
          "André Gide"
        ]
      },
      "id": "1058",
      "score": 1.8723381430998982,
      "timestamp": 1644936891229208800
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
      "highlights": {
        "text": [
          "…o derived from the ship Argo in Greek mythology, sailed by Jason and the Argonauts to Colchis in <mark>search</mark> of the Golden Fleece. Due to precession of the equinoxes, the position of the stars from Earth&#39;s…"
        ],
        "title": [
          "Argo Navis"
        ]
      },
      "id": "1924",
      "score": 1.8495595076783102,
      "timestamp": 1644936891230542600
    },
    {
      "fields": {
        "text": [
          "Asia () is Earth's largest and most populous continent, located primarily in the Eastern and Northern Hemispheres. It shares the continental landmass of Eurasia with the continent of Europe and the continental landmass of Afro-Eurasia with both Europe and Africa. Asia covers an area of , about 30% of Earth's total land area and 8.7% of the Earth's total surface area. The continent, which has long been home to the majority of the human population, was the site of many of the first civilizations. Its 4.5 billion people () constitute roughly 60% of the world's population.\nIn general terms, Asia is bounded on the east by the Pacific Ocean, on the south by the Indian Ocean, and on the north by the Arctic Ocean. The border of Asia with Europe is a historical and cultural construct, as there is no clear physical and geographical separation between them. It is somewhat arbitrary and has moved since its first conception in classical antiquity. The division of Eurasia into two continents reflects East–West cultural, linguistic, and ethnic differences, some of which vary on a spectrum rather than with a sharp dividing line. The most commonly accepted boundaries place Asia to the east of the Suez Canal separating it from Africa; and to the east of the Turkish Straits, the Ural Mountains and Ural River, and to the south of the Caucasus Mountains and the Caspian and Black Seas, separating it from Europe.\nChina and India alternated in being the largest economies in the world from 1 to 1800 CE. China was a major economic power and attracted many to the east, and for many the legendary wealth and prosperity of the ancient culture of India personified Asia, attracting European commerce, exploration and colonialism. The accidental discovery of a trans-Atlantic route from Europe to America by Columbus while in search for a route to India demonstrates this deep fascination. The Silk Road became the main east–west trading route in the Asian hinterlands while the Straits of Malacca stood as a major sea route. Asia has exhibited economic dynamism (particularly East Asia) as well as robust population growth during the 20th century, but overall population growth has since fallen. Asia was the birthplace of most of the world's mainstream religions including Hinduism, Zoroastrianism, Judaism, Jainism, Buddhism, Confucianism, Taoism, Christianity, Islam, Sikhism, as well as many other religions.\nGiven its size and diversity, the concept of Asia—a name dating back to classical antiquity—may actually have more to do with human geography than physical geography. Asia varies greatly across and within its regions with regard to ethnic groups, cultures, environments, economics, historical ties and government systems. It also has a mix of many different climates ranging from the equatorial south via the hot desert in the Middle East, temperate areas in the east and the continental centre to vast subarctic and polar areas in Siberia."
        ],
        "title": [
          "Asia"
        ],
        "url": [
          "https://en.wikipedia.org/wiki/Asia"
        ]
      },
      "highlights": {
        "text": [
          "…sm. The accidental discovery of a trans-Atlantic route from Europe to America by Columbus while i<mark>n sear</mark>ch for a route to India demonstrates this deep fascination. The Silk Road became the main east–we…"
        ],
        "title": [
          "Asia"
        ]
      },
      "id": "689",
      "score": 1.085247182156043,
      "timestamp": 1644936891216916000
    },
    {
      "fields": {
        "text": [
          "Artificial intelligence (AI) is intelligence demonstrated by machines, as opposed to natural intelligence displayed by animals including humans.\nLeading AI textbooks define the field as the study of \"intelligent agents\": any system that perceives its environment and takes actions that maximize its chance of achieving its goals.\nSome popular accounts use the term \"artificial intelligence\" to describe machines that mimic \"cognitive\" functions that humans associate with the human mind, such as \"learning\" and \"problem solving\", however, this definition is rejected by major AI researchers.\nAI applications include advanced web search engines (e.g., Google), recommendation systems (used by YouTube, Amazon and Netflix), understanding human speech (such as Siri and Alexa), self-driving cars (e.g., Tesla), automated decision-making and competing at the highest level in strategic game systems (such as chess and Go).\nAs machines become increasingly capable, tasks considered to require \"intelligence\" are often removed from the definition of AI, a phenomenon known as the AI effect. For instance, optical character recognition is frequently excluded from things considered to be AI, having become a routine technology.\nArtificial intelligence was founded as an academic discipline in 1956, and in the years since has experienced several waves of optimism,\nand have been common in fiction, as in Mary Shelley's Frankenstein or Karel Čapek's R.U.R. These characters and their fates raised many of the same issues now discussed in the ethics of artificial intelligence.\nThe study of mechanical or \"formal\" reasoning began with philosophers and mathematicians in antiquity. The study of mathematical logic led directly to Alan Turing's theory of computation, which suggested that a machine, by shuffling symbols as simple as \"0\" and \"1\", could simulate any conceivable act of mathematical deduction. This insight that digital computers can simulate any process of formal reasoning is known as the Church–Turing thesis.\nThe Church-Turing thesis, along with concurrent discoveries in neurobiology, information theory and cybernetics, led researchers to consider the possibility of building an electronic brain.\nThe first work that is now generally recognized as AI was McCullouch and Pitts' 1943 formal design for Turing-complete \"artificial neurons\".\nWhen access to digital computers became possible in the mid-1950s, AI research began to explore the possibility that human intelligence could be reduced to step-by-step symbol manipulation, known as Symbolic AI or GOFAI. Approaches based on cybernetics or artificial neural networks were abandoned or pushed into the background.\nThe field of AI research was born at a workshop at Dartmouth College in 1956.\nThe attendees became the founders and leaders of AI research.\nThey and their students produced programs that the press described as \"astonishing\":\ncomputers were learning checkers strategies, solving word problems in algebra, proving logical theorems and speaking English.\nBy the middle of the 1960s, research in the U.S. was heavily funded by the Department of Defense\nand laboratories had been established around the world.\nResearchers in the 1960s and the 1970s were convinced that symbolic approaches would eventually succeed in creating a machine with artificial general intelligence and considered this the goal of their field.\nHerbert Simon predicted, \"machines will be capable, within twenty years, of doing any work a man can do\".\nMarvin Minsky agreed, writing, \"within a generation ... the problem of creating 'artificial intelligence' will substantially be solved\".\nThey failed to recognize the difficulty of some of the remaining tasks. Progress slowed and in 1974, in response to the criticism of Sir James Lighthill\nand ongoing pressure from the US Congress to fund more productive projects, both the U.S. and British governments cut off exploratory research in AI. The next few years would later be called an \"AI winter\", a period when obtaining funding for AI projects was difficult.\nIn the early 1980s, AI research was revived by the commercial success of expert systems,\na form of AI program that simulated the knowledge and analytical skills of human experts. By 1985, the market for AI had reached over a billion dollars. At the same time, Japan's fifth generation computer project inspired the U.S and British governments to restore funding for academic research.\nHowever, beginning with the collapse of the Lisp Machine market in 1987, AI once again fell into disrepute, and a second, longer-lasting winter began.\nMany researchers began to doubt that the symbolic approach would be able to imitate all the processes of human cognition, especially perception, robotics, learning and pattern recognition. A number of researchers began to look into \"sub-symbolic\" approaches to specific AI problems. Robotics researchers, such as Rodney Brooks, rejected symbolic AI and focused on the basic engineering problems that would allow robots to move, survive, and learn their environment.\nInterest in neural networks and \"connectionism\" was revived by Geoffrey Hinton, David Rumelhart and others in the middle of the 1980s.\nSoft computing tools were developed in the 80s, such as neural networks, fuzzy systems, Grey system theory, evolutionary computation and many tools drawn from statistics or mathematical optimization.\nAI gradually restored its reputation in the late 1990s and early 21st century by finding specific solutions to specific problems. The narrow focus allowed researchers to produce verifiable results, exploit more mathematical methods, and collaborate with other fields (such as statistics, economics and mathematics).\nBy 2000, solutions developed by AI researchers were being widely used, although in the 1990s they were rarely described as \"artificial intelligence\".\nFaster computers, algorithmic improvements, and access to large amounts of data enabled advances in machine learning and perception; data-hungry deep learning methods started to dominate accuracy benchmarks around 2012.\nAccording to Bloomberg's Jack Clark, 2015 was a landmark year for artificial intelligence, with the number of software projects that use AI within Google increased from a \"sporadic usage\" in 2012 to more than 2,700 projects. He attributes this to an increase in affordable neural networks, due to a rise in cloud computing infrastructure and to an increase in research tools and datasets. In a 2017 survey, one in five companies reported they had \"incorporated AI in some offerings or processes\". The amount of research into AI (measured by total publications) increased by 50% in the years 2015–2019.\nNumerous academic researchers became concerned that AI was no longer pursuing the original goal of creating versatile, fully intelligent machines. Much of current research involves statistical AI, which is overwhelmingly used to solve specific problems, even highly successful techniques such as deep learning. This concern has led to the subfield artificial general intelligence (or \"AGI\"), which had several well-funded institutions by the 2010s."
        ],
        "title": [
          "Artificial intelligence"
        ],
        "url": [
          "https://en.wikipedia.org/wiki/Artificial_intelligence"
        ]
      },
      "highlights": {
        "text": [
          "…wever, this definition is rejected by major AI researchers.\nAI applications include advanced web <mark>search</mark> engines (e.g., Google), recommendation systems (used by YouTube, Amazon and Netflix), understand…"
        ],
        "title": [
          "Artificial intelligence"
        ]
      },
      "id": "1164",
      "score": 0.5769028632224984,
      "timestamp": 1644936891228777200
    }
  ],
  "hits": 6,
  "index_name": "enwiki"
}
```
