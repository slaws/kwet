package main

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},
	Route{
		"TodoIndex",
		"GET",
		"/todos",
		TodoIndex,
	},
	Route{
		"QList",
		"GET",
		"/q",
		ListQueue,
	},
	Route{
		"AddEvent",
		"POST",
		"/event",
		PostEvent,
	},
}
