package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/mijia/sweb/log"
)

const (
	kAssetsReverseKey = "_!#assets_"
)

// EnableExtraAssetsMapping can be used to set some extra assets mapping data for server,
// server would look up this mapping for the frontend assets first when reverse an assets url.
func (s *Server) EnableExtraAssetsMapping(assetsMapping map[string]string) {
	s.extraAssetsMapping = assetsMapping
}

func (s *Server) EnableExtraAssetsJson(jsonFile string) {
	s.extraAssetsJson = jsonFile
	s.loadJsonAssetsMapping()
}

// EnableAssetsPrefix can be used to add assets prefix for assets reverse, like CDN host name.
func (s *Server) EnableAssetsPrefix(prefix string) {
	s.assetsPrefix = prefix
}

func (s *Server) loadJsonAssetsMapping() {
	if s.extraAssetsJson == "" {
		return
	}
	if data, err := ioutil.ReadFile(s.extraAssetsJson); err == nil {
		mapping := make(map[string]string)
		if err := json.Unmarshal(data, &mapping); err == nil {
			s.extraAssetsMapping = mapping
			log.Debugf("Server extra assets loaded from json file: %s", s.extraAssetsJson)
		}
	}
}

// Reverse would reverse the named routes with params supported. E.g. we have a routes "/hello/:name" named "Hello",
// then we can call s.Reverse("Hello", "world") gives us "/hello/world"
func (s *Server) Reverse(name string, params ...interface{}) string {
	path, ok := s.namedRoutes[name]
	if !ok {
		log.Warnf("Server routes reverse failed, cannot find named routes %q", name)
		return "/no_such_named_routes_defined"
	}
	if len(params) == 0 || path == "/" {
		return path
	}
	strParams := make([]string, len(params))
	for i, param := range params {
		strParams[i] = fmt.Sprint(param)
	}
	parts := strings.Split(path, "/")[1:]
	paramIndex := 0
	for i, part := range parts {
		if part[0] == ':' || part[0] == '*' {
			if paramIndex < len(strParams) {
				parts[i] = strParams[paramIndex]
				paramIndex++
			}
		}
	}
	return httprouter.CleanPath("/" + strings.Join(parts, "/"))
}

// Assets would reverse the assets url, e.g. s.Assets("images/test.png") gives us "/assets/images/test.png"
func (s *Server) Assets(path string) string {
	if asset, ok := s.extraAssetsMapping[path]; ok {
		path = asset
	}
	return fmt.Sprintf("%s%s", s.assetsPrefix, s.Reverse(kAssetsReverseKey, path))
}

// Files register the static file system or assets to the router
func (s *Server) Files(path string, root http.FileSystem) {
	s.namedRoutes[kAssetsReverseKey] = path
	s.router.ServeFiles(path, root)
}

// Use FileHandlerHook to alter Response header or something
type FileHandlerHook func(w http.ResponseWriter, r *http.Request)

func (s *Server) FilesWithHook(path string, root http.FileSystem, hook FileHandlerHook) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	s.namedRoutes[kAssetsReverseKey] = path

	fileServer := http.FileServer(root)
	s.router.GET(path, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		r.URL.Path = ps.ByName("filepath")
		hook(w, r)
		fileServer.ServeHTTP(w, r)
	})
}
