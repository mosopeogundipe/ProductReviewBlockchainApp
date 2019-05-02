package p5

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var userRoutes = Routes{
	Route{
		"RegisterUser",
		"GET",
		"/user/register",
		RegisterUser,
	},
	Route{
		"RegisterProduct",
		"POST",
		"/product/register",
		RegisterProduct,
	},
	Route{
		"RegisterMiner",
		"GET",
		"/miner/register",
		RegisterMiner,
	},
	Route{
		"PostReview",
		"POST",
		"/review/post",
		PostReview,
	},
	Route{
		"SignMessage",
		"POST",
		"/sign/message",
		SignMessage,
	},
}
