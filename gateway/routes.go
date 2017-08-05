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
	Route{
		"EditHubRule",
		"GET",
		"/settings/hub/rule/{rulename}",
		EditHubRule,
	},
	Route{
		"UpdateHubRule",
		"POST",
		"/settings/hub/rule/{rulename}",
		EditHubRule,
	},
	Route{
		"DeleteHubRule",
		"DELETE",
		"/settings/hub/rule/{rulename}",
		DeleteHubRule,
	},
	Route{
		"AddNewNotifRule",
		"GET",
		"/notif/new",
		AddNewNotifRule,
	},
	Route{
		"AddNewNotifRule",
		"POST",
		"/notif/new",
		AddNewNotifRule,
	},
	Route{
		"UpdateNotifRule",
		"GET",
		"/settings/notifier/rule/{rulename}",
		EditNotifRule,
	},
	Route{
		"UpdateNotifRule",
		"POST",
		"/settings/notifier/rule/{rulename}",
		EditNotifRule,
	},
	Route{
		"DeleteNotifRule",
		"DELETE",
		"/settings/notifier/rule/{rulename}",
		DeleteNotifRule,
	},
}
