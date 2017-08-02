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
		"AddEvent",
		"POST",
		"/event/{application}",
		PostEvent,
	},
	Route{
		"WS",
		"GET",
		"/wevents",
		SocketEvent,
	},
	// Route{
	// 	"WS",
	// 	"GET",
	// 	"/wevents/{queue:[a-zA-Z0-9\\.\\-\\_@]}",
	// 	SocketSpecificEvent,
	// },
	Route{
		"QueueList",
		"GET",
		"/queues",
		ListQueues,
	},
	// Route{
	// 	"Static",
	// 	"GET",
	// 	"/static/",
	// 	ServeStatic,
	// },
}
