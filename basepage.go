package main

import (
	"net/http"
	"io/ioutil"
	"strings"
	"path/filepath"
	"log"
	"html/template"
)

func serveBaseHTML(w http.ResponseWriter, r *http.Request) {
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
				{{range $i, $e := .}}<li>
				    <a href="{{.Url}}">{{.Title}}</a>
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

	files, err := ioutil.ReadDir("webroot/configs")
	if err != nil {
		log.Fatal(err)
	}

	links := []Link{}
	for _, file := range files {
		config := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		title := strings.Title(strings.Replace(strings.Replace(config, "-", " ", -1), "_", " ", -1))
		links = append(links, Link{
			Url:   "?config=" + config,
			Title: title,
		})
	}

	templ, err := template.New("base").Parse(baseHTML)
	if err != nil {
		log.Fatal(err)
	}

	err = templ.Execute(w, links)
	if err != nil {
		log.Fatal(err)
	}
}

