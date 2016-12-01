package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

type Link struct {
	URL   string
	Title string
}

func serveBaseHTML(template *template.Template, bookie bookie, w http.ResponseWriter, r *http.Request) error {
	files, err := ioutil.ReadDir("webroot/configs")
	if err != nil {
		return err
	}

	ids := []string{}
	if fsms, err := bookie.latestFSMs(10); err == nil {
		for _, f := range fsms {
			ids = append(ids, f.Id)
		}
	}

	links := []Link{}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".js" {
			config := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			title := strings.Title(strings.Replace(strings.Replace(config, "-", " ", -1), "_", " ", -1))
			links = append(links, Link{
				URL:   "?config=" + config,
				Title: title,
			})
			for _, d := range ids {
				links = append(links, Link{
					URL:   "?config=" + config + "&fsmId=" + d,
					Title: title + " > " + d,
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
				    <a href="{{.URL}}">{{.Title}}</a>
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
			</style>
			</body>
			</html>
	`

	return template.New("base").Parse(baseHTML)
}
