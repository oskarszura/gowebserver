package router

import (
	"github.com/coda-it/gowebserver/session"
	"github.com/coda-it/gowebserver/store"
	"github.com/coda-it/gowebserver/utils/logger"
	"github.com/coda-it/gowebserver/utils/url"
	"net/http"
	"regexp"
	"strings"
)

// IRouter - router interface
type IRouter interface {
	Route(w http.ResponseWriter, r *http.Request)
	AddRoute(w http.ResponseWriter, r *http.Request)
}

// Router - router struct
type Router struct {
	sessionManager         session.ISessionManager
	urlRoutes              []URLRoute
	pageNotFoundController ControllerHandler
	store                  store.IStore
}

// New - factory for router
func New(sm session.SessionManager, notFound ControllerHandler) Router {
	return Router{
		sessionManager:         sm,
		urlRoutes:              make([]URLRoute, 0),
		pageNotFoundController: notFound,
		store:                  store.New(),
	}
}

func (router Router) findRoute(path string, method string) URLRoute {
	for _, v := range router.urlRoutes {
		pathRegExp := regexp.MustCompile(v.urlRegExp)

		if pathRegExp.MatchString(path) && (v.method == method || v.method == "ALL") {
			return v
		}
	}
	return URLRoute{
		handler: router.pageNotFoundController,
	}
}

// New - factory for session manager
func (router *Router) New(sm session.ISessionManager) {
	router.sessionManager = sm
}

// Route - routes all incomming requests
func (router *Router) Route(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	route := router.findRoute(urlPath, r.Method)

	params := make(map[string]string)
	pathItems := strings.Split(urlPath, "/")

	for k, v := range route.params {
		if len(pathItems) > v {
			params[k] = pathItems[v]
		}
	}

	urlOptions := &URLOptions{
		params,
	}

	logger.Log(logger.INFO, "Navigating to url = "+urlPath+" vs route = "+
		route.urlRegExp)

	routeHandler := route.handler
	routeHandler(w, r, *urlOptions, router.sessionManager, router.store)
}

// AddRoute - adds route
func (router *Router) AddRoute(urlPattern string, method string, pathHandler ControllerHandler) {
	params := make(map[string]int)
	pathRegExp := url.UrlPatternToRegExp(urlPattern)

	urlPathItems := strings.Split(urlPattern, "/")

	for i := 0; i < len(urlPathItems); i++ {
		paramKey := urlPathItems[i]
		isParam, _ := regexp.MatchString(`{[a-zA-Z0-9]*}`, paramKey)

		if isParam {
			strippedParamKey := strings.Replace(strings.Replace(paramKey,
				"{", "", -1), "}", "", -1)
			params[strippedParamKey] = i
		}
	}

	router.urlRoutes = append(router.urlRoutes, URLRoute{
		urlRegExp: pathRegExp,
		method:    method,
		handler:   pathHandler,
		params:    params,
	})
}

// AddDataSource - adds data source
func (router *Router) AddDataSource(name string, ds interface{}) {
	router.store.AddDataSource(name, ds)
}
