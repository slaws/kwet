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
	Route{
		"QueueList",
		"GET",
		"/queues",
		ListQueues,
	},
	Route{
		"Settings",
		"GET",
		"/settings",
		Settings,
	},
	Route{
		"UpdateSettings",
		"POST",
		"/settings/{module}",
		UpdateSettings,
	},
	Route{
		"AddNewHubRule",
		"GET",
		"/addnewhubrule",
		AddNewHubRule,
	},
	Route{
		"AddNewHubRule",
		"POST",
		"/addnewhubrule",
		AddNewHubRule,
	},
}
