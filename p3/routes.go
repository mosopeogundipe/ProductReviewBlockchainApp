package p3

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
		"Show",
		"GET",
		"/show",
		Show,
	},
	Route{
		"Upload",
		"GET",
		"/upload",
		Upload,
	},
	Route{
		"UploadBlock",
		"GET",
		"/block/{height}/{hash}",
		UploadBlock,
	},
	Route{
		"HeartBeatReceive",
		"POST",
		"/heartbeat/receive",
		HeartBeatReceive,
	},
	Route{
		"Start",
		"GET",
		"/start",
		Start,
	},
	Route{
		"Canonical",
		"GET",
		"/canonical",
		Canonical,
	},
	Route{
		"TransactionReceive",
		"POST",
		"/transaction/receive",
		TransactionReceive,
	},
	Route{
		"FindReviewsByPublicKey",
		"POST",
		"/reviews/find/publickey",
		FindReviewsByPublicKey,
	},
	Route{
		"FindReviewsByProductAndUserID",
		"POST",
		"/reviews/find/all",
		FindReviewsByProductAndUserID,
	},
}
