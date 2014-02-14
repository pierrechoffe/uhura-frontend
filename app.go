package main

import (
	"code.google.com/p/goauth2/oauth"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/dukex/uhura/core"
	"github.com/rakyll/martini-contrib/cors"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var config oauth.Config
var homeChannel []core.ChannelResult

const profileInfoURL = "https://www.googleapis.com/oauth2/v1/userinfo?alt=json"

func emberAppHandler(r render.Render, req *http.Request) string {
	var indexTemplate string
	userAgent := strings.ToLower(req.UserAgent())

	fmt.Println("USER AGENT", userAgent)

	if strings.Contains(userAgent, "flipboard.com") || strings.Contains(userAgent, "newsme") || strings.Contains(userAgent, "bot") || strings.Contains(userAgent, "slurp") || strings.Contains(userAgent, "facebookexternalhit") {
		url := os.Getenv("PRERENDER_SERVER") + "/" + "http://" + req.Host + req.URL.RequestURI()
		fmt.Println(url)

		res, err := http.Get(url)
		defer res.Body.Close()
		if err != nil {
			fmt.Println("ERROR 1", err)
		}

		body, err := ioutil.ReadAll(res.Body)

		if err != nil {
			fmt.Println("ERROR 2", err)
		}
		indexTemplate = string(body)
	} else {
		baseUrl := "http://uhuraapp.com"
		if os.Getenv("ENV") == "development" {
			baseUrl = "http://127.0.0.1:3002"
		}

		itb, _ := ioutil.ReadFile("./templates/index.html")
		indexTemplate = string(itb[:])
		indexTemplate = strings.Replace(indexTemplate, "<% URL %>", baseUrl, -1)
	}

	return indexTemplate
}

func main() {
	sids := strings.Split(os.Getenv("CHANNELS_HOME"), ",")
	var ids = make([]int, 0)

	for i, _ := range sids {
		id, _ := strconv.Atoi(sids[i])
		ids = append(ids, id)
	}
	homeChannel = core.GetChannels(ids)

	config := &oauth.Config{
		ClientId:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
		Scope:        "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile",
		AuthURL:      "https://accounts.google.com/o/oauth2/auth",
		TokenURL:     "https://accounts.google.com/o/oauth2/token",
		TokenCache:   oauth.CacheFile("cache.json"),
	}

	m := martini.Classic()

	m.Use(render.Renderer(render.Options{
		Funcs: []template.FuncMap{
			{
				"UserViewedHelper": core.UserViewedHelper,
				"Pagination":       core.Pagination,
			},
		},
	}))

	m.Use(cors.Allow(&cors.Opts{
		AllowOrigins: []string{"http://emberjs.jsbin.com"},
	}))

	m.Use(martini.Static("assets"))
	m.Use(martini.Static("fonts"))
	m.Use(martini.Recovery())
	m.Use(core.UhuraRecovery())

	// API

	m.Post("/api/channels/:id/fetcher", func(r render.Render, params martini.Params) {
		paramsId, _ := strconv.Atoi(params["id"])
		channel := core.GetChannel(paramsId)
		core.FetchChanell(channel)
		r.JSON(202, nil)
	})

	m.Post("/api/episodes/:id/listened", func(responseWriter http.ResponseWriter, r render.Render, request *http.Request, params martini.Params) {
		user, err := core.CurrentUser(request)
		if err {
			r.Error(403)
			return
		}

		episode := core.UserListen(user.Id, params["id"])
		r.JSON(202, map[string]interface{}{"episode": episode})
	})

	// -----------------
	// API
	m.Get("/api/channels", func(r render.Render, w http.ResponseWriter, request *http.Request) {
		var channels []core.ChannelResult

		if featured := request.FormValue("featured"); featured == "" {
			var userId int
			user, err := core.CurrentUser(request)
			if err {
				userId = 0
			} else {
				userId = user.Id
			}

			channels, _ = core.AllChannels(userId, false, 0)
		} else {
			channels = homeChannel
		}

		r.JSON(200, map[string]interface{}{"channels": channels})
	})

	m.Post("/api/channels", func(responseWriter http.ResponseWriter, r render.Render, request *http.Request, params martini.Params) {
		user, err := core.CurrentUser(request)
		if err {
			r.Error(403)
			return
		}

		var channelJson struct {
			Channel core.Channel `json:"channel"`
		}

		json.NewDecoder(request.Body).Decode(&channelJson)

		channel := core.AddFeed(channelJson.Channel.Url, user.Id)
		r.JSON(200, map[string]interface{}{"channel": channel})
	})

	m.Get("/api/channels/:id", func(r render.Render, params martini.Params, request *http.Request) {
		var userId int
		user, err := core.CurrentUser(request)
		if err {
			userId = 0
		} else {
			userId = user.Id
		}

		channelId, _ := strconv.Atoi(params["id"])

		channels, episodes := core.AllChannels(userId, false, channelId)

		r.JSON(200, map[string]interface{}{"channel": channels[0], "episodes": episodes})
	})

	m.Get("/api/channels/:id/subscribe", func(r render.Render, request *http.Request, params martini.Params) {
		user, err := core.CurrentUser(request)

		if err {
			r.Error(403)
			return
		}

		channel := core.SubscribeChannel(user.Id, params["id"])

		r.JSON(200, map[string]interface{}{"channel": channel})
	})

	m.Delete("/api/channels/:id/subscribe", func(r render.Render, request *http.Request, params martini.Params) {
		user, err := core.CurrentUser(request)

		if err {
			r.Error(403)
			return
		}

		channel := core.UnsubscribeChannel(user.Id, params["id"])

		r.JSON(200, map[string]interface{}{"channel": channel})
	})

	// API
	m.Get("/api/subscriptions", func(r render.Render, w http.ResponseWriter, request *http.Request) {
		user, err := core.CurrentUser(request)
		if err {
			r.Error(403)
			return
		}
		subscribes, channels := core.Subscriptions(user)
		r.JSON(200, map[string]interface{}{"subscriptions": subscribes, "channels": channels})
	})

	m.Get("/api/episodes", func(request *http.Request, r render.Render) {
		ids := request.URL.Query()["ids[]"]
		episodes := core.GetItems(ids)
		r.JSON(200, map[string]interface{}{"episodes": episodes})
	})

	m.Get("/api/episodes/:id", func(params martini.Params, request *http.Request, r render.Render) {
		user, _ := core.CurrentUser(request)

		idInt, _ := strconv.Atoi(params["id"])
		episode, notFound := core.GetItem(idInt, user.Id)
		if notFound {
			r.JSON(404, nil)
		} else {
			r.JSON(200, map[string]interface{}{"episode": episode})
		}
	})

	m.Put("/api/episodes/:id", func(params martini.Params, request *http.Request, r render.Render) {
		user, err := core.CurrentUser(request)
		if err {
			r.Error(403)
			return
		}

		idInt, _ := strconv.Atoi(params["id"])
		episode, _ := core.GetItem(idInt, user.Id)

		r.JSON(200, map[string]interface{}{"episode": episode})

	})

	m.Get("/api/subscriptions/:id/episodes", func(r render.Render, params martini.Params, request *http.Request) {
		user, err := core.CurrentUser(request)
		if err {
			r.Error(403)
			return
		}

		idInt, _ := strconv.Atoi(params["id"])

		channels, episodes := core.AllChannels(user.Id, false, idInt)
		r.JSON(200, map[string]interface{}{"episodes": episodes, "channel": channels[0]})
	})

	// API - Auth
	m.Get("/api/authorize", func(w http.ResponseWriter, request *http.Request) string {
		_, err := core.CurrentUser(request)
		if err {
			url := config.AuthCodeURL("")
			http.Redirect(w, request, url, http.StatusFound)
			return ""
		} else {
			return "<script>window.close();</script>"
		}
	})

	m.Get("/auth/callback", func(responseWriter http.ResponseWriter, request *http.Request) string {
		code := request.FormValue("code")
		t := &oauth.Transport{Config: config}
		t.Exchange(code)
		responseAuth, _ := t.Client().Get(profileInfoURL)
		defer responseAuth.Body.Close()

		core.CreateAndLoginUser(request, responseWriter, responseAuth)

		return "<script>window.close();</script>"
	})

	// API - DEV
	m.Get("/api/dev/fetchall", func(r render.Render) {
		if os.Getenv("ENV") == "development" {
			core.FetchAllChannell()
		}
		r.JSON(202, "")
	})

	m.Get("/sitemap.xml", func() string {
		return core.SiteMap()
	})

	m.Get("/**", emberAppHandler)

	fmt.Println("Starting server on", os.Getenv("PORT"))
	err := http.ListenAndServe(":"+os.Getenv("PORT"), m)
	if err != nil {
		panic(err)
	}
}
