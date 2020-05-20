package jsonquery

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseString(s string) (*Node, error) {
	return Parse(strings.NewReader(s))
}

func TestParseJsonNumberArray(t *testing.T) {
	s := `[1,2,3,4,5,6]`
	doc, err := parseString(s)
	if err != nil {
		t.Fatal(err)
	}
	// output like below:
	// <element>1</element>
	// <element>2</element>
	// ...
	// <element>6</element>
	if e, g := 6, len(doc.ChildNodes()); e != g {
		t.Fatalf("excepted %v but got %v", e, g)
	}
	var v []string
	for _, n := range doc.ChildNodes() {
		v = append(v, n.InnerText())
	}
	if got, expected := strings.Join(v, ","), "1,2,3,4,5,6"; got != expected {
		t.Fatalf("got %v but expected %v", got, expected)
	}
}

func TestParseJsonObject(t *testing.T) {
	s := `{
		"name":"John",
		"age":31,
		"city":"New York"
	}`
	doc, err := parseString(s)
	if err != nil {
		t.Fatal(err)
	}
	// output like below:
	// <name>John</name>
	// <age>31</age>
	// <city>New York</city>
	m := make(map[string]string)
	for _, n := range doc.ChildNodes() {
		m[n.Data] = n.InnerText()
	}
	expected := []struct {
		name, value string
	}{
		{"name", "John"},
		{"age", "31"},
		{"city", "New York"},
	}
	for _, v := range expected {
		if e, g := v.value, m[v.name]; e != g {
			t.Fatalf("expected %v=%v,but %v=%v", v.name, e, v.name, g)
		}
	}
}

func TestParseJsonObjectArray(t *testing.T) {
	s := `[
		{ "name":"Ford", "models":[ "Fiesta", "Focus", "Mustang" ] },
		{ "name":"BMW", "models":[ "320", "X3", "X5" ] },
        { "name":"Fiat", "models":[ "500", "Panda" ] }
	]`
	doc, err := parseString(s)
	if err != nil {
		t.Fatal(err)
	}
	/**
	<element>
		<name>Ford</name>
		<models>
			<element>Fiesta</element>
			<element>Focus</element>
			<element>Mustang</element>
		</models>
	</element>
	<element>
		<name>BMW</name>
		<models>
			<element>320</element>
			<element>X3</element>
			<element>X5</element>
		</models>
	</element>
	....
	*/
	if e, g := 3, len(doc.ChildNodes()); e != g {
		t.Fatalf("expected %v, but %v", e, g)
	}
	m := make(map[string][]string)
	for _, n := range doc.ChildNodes() {
		// Go to the next of the element list.
		var name string
		var models []string
		for _, e := range n.ChildNodes() {
			if e.Data == "name" {
				// a name node.
				name = e.InnerText()
			} else {
				// a models node.
				for _, k := range e.ChildNodes() {
					models = append(models, k.InnerText())
				}
			}
		}
		// Sort models list.
		sort.Strings(models)
		m[name] = models

	}
	expected := []struct {
		name, value string
	}{
		{"Ford", "Fiesta,Focus,Mustang"},
		{"BMW", "320,X3,X5"},
		{"Fiat", "500,Panda"},
	}
	for _, v := range expected {
		if e, g := v.value, strings.Join(m[v.name], ","); e != g {
			t.Fatalf("expected %v=%v,but %v=%v", v.name, e, v.name, g)
		}
	}
}

func TestParseJson(t *testing.T) {
	s := `{
		"name":"John",
		"age":30,
		"cars": [
			{ "name":"Ford", "models":[ "Fiesta", "Focus", "Mustang" ] },
			{ "name":"BMW", "models":[ "320", "X3", "X5" ] },
			{ "name":"Fiat", "models":[ "500", "Panda" ] }
		]
	 }`
	doc, err := parseString(s)
	if err != nil {
		t.Fatal(err)
	}
	n := doc.SelectElement("name")
	if n == nil {
		t.Fatal("n is nil")
	}
	if n.NextSibling != nil {
		t.Fatal("next sibling shoud be nil")
	}
	if e, g := "John", n.InnerText(); e != g {
		t.Fatalf("expected %v but %v", e, g)
	}
	cars := doc.SelectElement("cars")
	if e, g := 3, len(cars.ChildNodes()); e != g {
		t.Fatalf("expected %v but %v", e, g)
	}
}

func TestLargeFloat(t *testing.T) {
	s := `{
		"large_number": 365823929453
	 }`
	doc, err := parseString(s)
	if err != nil {
		t.Fatal(err)
	}
	n := doc.SelectElement("large_number")
	if n.InnerText() != "365823929453" {
		t.Fatalf("expected %v but %v", "365823929453", n.InnerText())
	}
}

func TestConvert(t *testing.T) {
	config := `
{
    "top" : {
	"inner" : [ 0,1,2,3 ],
	"people" : [
	    {
		"name": "joe",
		"age": 45
	    },       {
		"name": "mark",
		"age": 2
	    }
	],
	"route-instance" : {
             "ri1" : {
                "metric" : 24
             },
             "ri2" : {
                "metric" : 89
             }
       }
    }
}
`
	jtree := map[string]interface{}{}
	err := json.Unmarshal([]byte(config), &jtree)
	assert.Nil(t, err)

	doc := ParseTree(jtree)

	tree := ConvertNodeToInterface(doc)

	outbytes, err := json.MarshalIndent(tree, "", "  ")
	assert.Nil(t, err)

	exp := `{
  "top": {
    "inner": [
      "0",
      "1",
      "2",
      "3"
    ],
    "people": [
      {
        "age": "45",
        "name": "joe"
      },
      {
        "age": "2",
        "name": "mark"
      }
    ],
    "route-instance": {
      "ri1": {
        "metric": "24"
      },
      "ri2": {
        "metric": "89"
      }
    }
  }
}`
	assert.Equal(t, string(outbytes), exp)
}

func TestQueryConvert(t *testing.T) {
	config := `
{
    "top" : {
	"inner" : [ 0,1,2,3 ],
	"people" : [
	    {
		"name": "joe",
		"age": 45
	    },       {
		"name": "mark",
		"age": 2
	    }
	],
	"route-instance" : {
            "ri1" : {
                "metric" : 24
            },
            "ri2" : {
                "metric" : 89
            }
	},
	"sites" : [
	    {
		"ri1" : {
		    "ospf" : {
			"areas" :  [
			    {
				"area_id" : "0.0.0.0",
				"metric" : 0
			    }
			]
		    }
		},
		"ri2" : {
		    "ospf" : {
			"areas": [
			    {
				"area_id" : "0.0.0.1",
				"metric" : 1
			    }
			]
		    }
		},
		"ri3" : {
		    "ospf" : {
			"areas" : [
			    {
				"area_id" : "0.0.0.2",
				"metric" : 2
			    }
			]
		    }
		}
	    }
	]
    }
}
`
	queryInOutExp := func(t *testing.T, config, query, exp string, fullPath bool) {
		t.Helper()

		doc, err := parseString(config)
		assert.Nil(t, err)

		names, err := QueryAll(doc, query)
		if err != nil {
			t.Error(err)
		}

		dst := ConvertNodesToInterface(names, fullPath)
		outbytes, err := json.MarshalIndent(dst, "", "  ")
		assert.Nil(t, err)

		assert.Equal(t, exp, string(outbytes))
	}

	exp := `[
  "joe",
  "mark"
]`
	queryInOutExp(t, config, "//name", exp, false)

	exp = `[
  {
    "top": {
      "people": [
        {
          "name": "joe"
        }
      ]
    }
  },
  {
    "top": {
      "people": [
        {
          "name": "mark"
        }
      ]
    }
  }
]`
	queryInOutExp(t, config, "//name", exp, true)

	exp = `[
  {
    "age": "2",
    "name": "mark"
  }
]`
	queryInOutExp(t, config, "//people/*[age < 44]", exp, false)

	exp = `[
  {
    "top": {
      "people": [
        {
          "age": "2",
          "name": "mark"
        }
      ]
    }
  }
]`
	queryInOutExp(t, config, "//people/*[age < 44]", exp, true)

	exp = `[
  {
    "metric": "24"
  }
]`
	queryInOutExp(t, config, "//route-instance/*[metric < 44]", exp, false)

	exp = `[
  {
    "top": {
      "route-instance": {
        "ri1": {
          "metric": "24"
        }
      }
    }
  }
]`
	queryInOutExp(t, config, "//route-instance/*[metric < 44]", exp, true)

	exp = `[
  {
    "area_id": "0.0.0.0",
    "metric": "0"
  },
  {
    "area_id": "0.0.0.2",
    "metric": "2"
  }
]`
	queryInOutExp(t, config, `//sites/*//*[area_id != "0.0.0.1"]`, exp, false)

	exp = `[
  {
    "top": {
      "sites": [
        {
          "ri1": {
            "ospf": {
              "areas": [
                {
                  "area_id": "0.0.0.0",
                  "metric": "0"
                }
              ]
            }
          }
        }
      ]
    }
  },
  {
    "top": {
      "sites": [
        {
          "ri3": {
            "ospf": {
              "areas": [
                {
                  "area_id": "0.0.0.2",
                  "metric": "2"
                }
              ]
            }
          }
        }
      ]
    }
  }
]`
	queryInOutExp(t, config, `//sites/*//*[area_id != "0.0.0.1"]`, exp, true)

}
