package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type Link struct {
	URL     string
	Title   string
	Tags    map[string]string
	Elapsed string
	Generic bool
}

func serveBaseHTML(template *template.Template, bookie bookie, w http.ResponseWriter, r *http.Request) error {
	files, err := ioutil.ReadDir("webroot/configs")
	if err != nil {
		return err
	}

	fsms := []fsm{}
	if f, err := bookie.latestFSMs(10); err == nil {
		fsms = f
	}

	links := []Link{}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".js" {
			config := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			title := strings.Title(strings.Replace(strings.Replace(config, "-", " ", -1), "_", " ", -1))
			links = append(links, Link{
				URL:     "?config=" + config,
				Title:   title,
				Generic: true,
			})
			l := len(fsms)
			for i := range fsms {
				fsm := fsms[l-i-1]
				elapsed := ""
				_created, err := time.Parse("2006-01-02T15:04:05Z", fsm.Created)
				if err == nil {
					elapsed = durationRound(time.Now().UTC().Sub(_created), time.Second).String()
				}

				links = append(links, Link{
					URL:     "?config=" + config + "&fsmId=" + fsm.Id,
					Title:   title + " > " + fsm.Id,
					Elapsed: elapsed + " ago",
					Tags:    fsm.Tags,
				})
			}
		}
	}

	return template.Execute(w, links)
}

func mustParseBasePageTemplate() *template.Template {
	template, err := parseBasePageTemplate()
	if err != nil {
		log.Fatalf("Could not parse base page template. err=%v", err)
	}
	return template
}

func parseBasePageTemplate() (*template.Template, error) {
	baseHTML := `
			<!DOCTYPE html>
			<html>
			<head>
			    <title>Flowbro</title>
			</head>
			<body>
			<div class="box">
			  <div class="row header">
			    Flowbro
			  </div>
			  <div class="row content">
			      <ul>
				{{range .}}<li>
				    <a href="{{.URL}}">{{.Title}}</a><br/>
				    {{if not .Generic }}<span class="created">(created {{.Elapsed}})</span>
				    {{range $key, $value := .Tags}}{{if ne $key "created"}}
							<br/>
							<span class="tag_key">{{$key}}</span>
							<span class="tag_value">{{$value}}</span>
				    {{end}}{{end}}{{end}}
				</li>{{end}}
			      </ul>
			  </div>
			  <div class="row footer">
			    <p><b>footer</b> (fixed height)</p>
			  </div>
			</div>
			<style>
				html,
				body {
				  background-color: #FFF;
			          margin: 0;
			          padding: 0;
			          line-height: 1;
			          font-family: 'Open Sans', 'Verdana', 'sans-serif';
                color: white;
				  height: 100%;
				}

				.box {
				  display: flex;
				  flex-flow: column;
				  height: 100%;
				}

				.box .row {
				  flex: 0 1 30px;
				}

				.box .row.header {
				  flex: 0 1 50px;
				  line-height: 50px;
				  font-size: 26px;
				  background-color: #0091EA;
				  padding: 20px;
				}

				.box .row.content {
				  flex: 1 1 auto;
				  color: black;
			    	  background-color: transparent;
			          box-shadow: inset 0px 3px 3px 1px rgba(0,0,0,0.3);
				}

				.box .row.footer {
				  flex: 0 1 40px;
				}

				ul {
				}

				li {
				  list-style: none;
				  padding-left:0;
				  padding: 20px;
				  font-size: 20px;
				}

				a {
				  text-decoration: none;
				  color: rgb(50, 50, 150);
				}

				a:hover {
				  color: rgb(50, 50, 200);
				}

				.created, .tag_key, .tag_value {
					padding-top: 5px;
					font-size: 12px;
					color: #777;
				}

				.tag_key {
					color: #222;
				}

				.tag_value {
				}
			</style>
			</body>
			</html>
	`

	return template.New("base").Parse(baseHTML)
}

// https://play.golang.org/p/QHocTHl8iR
func durationRound(d, r time.Duration) time.Duration {
	if r <= 0 {
		return d
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
}
