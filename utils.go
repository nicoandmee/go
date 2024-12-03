package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type urlProc func(*url.URL, string) []string
type UrlStruct struct {
	Scheme   string `json:"scheme"`
	Opaque   string `json:"opaque"`    // encoded opaque data scheme:opaque[?query][#fragment]
	User     string `json:"user"`      // username and password information
	Host     string `json:"host"`      // host or host:port  [scheme:][//[userinfo@]host][/]path[?query][#fragment]
	Path     string `json:"path"`      // path (relative paths may omit leading slash)
	RawPath  string `json:"raw_path"`  // encoded path hint (see EscapedPath method)
	RawQuery string `json:"raw_query"` // encoded query values, without '?'
	Fragment string `json:"fragment"`  // fragment for references, without '#'

	Parameters    []KeyValue `json:"parameters"` //
	Url           string     `json:"url"`        //
	Domain        string     `json:"domain"`     // The domain (e.g. sub.example.com)\n"
	Subdomain     string     `json:"subdomain"`  // The subdomain (e.g. sub)\n"
	Root          string     `json:"root"`       // The root of domain (e.g. example)\n"
	TLD           string     `json:"tld"`        // The TLD (e.g. com)\n"
	Apex          string     `json:"apex"`       //apex domain test.google.co.uk google.co.uk
	Port          string     `json:"port"`       // The port (e.g. 8080)\n"
	PathExtension string     `json:"extension"`  // The path's file extension (e.g. jpg, html)\n"
}

// parseURL parses a string as a URL and returns a *url.URL
// or any error that occured. If the initially parsed URL
// has no scheme, http:// is prepended and the string is
// re-parsed
func parseURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		return url.Parse("http://" + raw)
	}

	return u, nil
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func jsonFormat(u *url.URL, _ string) []string {
	parameters := make([]KeyValue, 0)
	for key, vals := range u.Query() {
		for _, val := range vals {
			parameters = append(parameters, KeyValue{Key: key, Value: val})
		}
	}
	extractApexs := format(u, "%r.%t")
	apex := ""
	if len(extractApexs) == 1 {
		apex = extractApexs[0]
	}
	domain := ""
	extractDomains := format(u, "%d")
	if len(extractDomains) == 1 {
		domain = extractDomains[0]
	}
	subdomain := ""
	extractSubdomains := format(u, "%S")
	if len(extractSubdomains) == 1 {
		subdomain = extractSubdomains[0]
	}
	root := ""
	extractRoots := format(u, "%r")
	if len(extractRoots) == 1 {
		root = extractRoots[0]
	}

	tld := ""
	extractTLDs := format(u, "%t")
	if len(extractTLDs) == 1 {
		tld = extractTLDs[0]
	}

	port := ""
	extractPorts := format(u, "%P")
	if len(extractPorts) == 1 {
		port = extractPorts[0]
	}

	extension := ""
	extractExtensions := format(u, "%e")
	if len(extractExtensions) == 1 {
		extension = extractExtensions[0]
	}

	newstructure := UrlStruct{
		Scheme:        u.Scheme,
		Opaque:        u.Opaque,
		User:          u.User.String(),
		Host:          u.Host,
		Path:          u.Path,
		RawPath:       u.RawPath,
		RawQuery:      u.RawQuery,
		Fragment:      u.Fragment,
		Parameters:    parameters,
		Apex:          apex,
		Url:           u.String(),
		Domain:        domain,
		Subdomain:     subdomain,
		Root:          root,
		TLD:           tld,
		Port:          port,
		PathExtension: extension,
	}
	outBytes, err := json.Marshal(newstructure)

	if err == nil {
		out := bytes.NewBuffer(outBytes).String()
		return []string{out}
	}
	return []string{""}
}

// keys returns all of the keys used in the query string
// portion of the URL. E.g. for /?one=1&two=2&three=3 it
// will return []string{"one", "two", "three"}
func keys(u *url.URL, _ string) []string {
	out := make([]string, 0)
	for key, _ := range u.Query() {
		out = append(out, key)
	}
	return out
}

// values returns all of the values in the query string
// portion of the URL. E.g. for /?one=1&two=2&three=3 it
// will return []string{"1", "2", "3"}
func values(u *url.URL, _ string) []string {
	out := make([]string, 0)
	for _, vals := range u.Query() {
		for _, val := range vals {
			out = append(out, val)
		}
	}
	return out
}

// keyPairs returns all the key=value pairs in
// the query string portion of the URL. E.g for
// /?one=1&two=2&three=3 it will return
// []string{"one=1", "two=2", "three=3"}
func keyPairs(u *url.URL, _ string) []string {
	out := make([]string, 0)
	for key, vals := range u.Query() {
		for _, val := range vals {
			out = append(out, fmt.Sprintf("%s=%s", key, val))
		}
	}
	return out
}

// domains returns the domain portion of the URL. e.g.
// for http://sub.example.com/path it will return
// []string{"sub.example.com"}
func domains(u *url.URL, f string) []string {
	return format(u, "%d")
}

// apexes return the apex portion of the URL. e.g.
// for http://sub.example.com/path it will return
// []string{"example.com"}
func apexes(u *url.URL, f string) []string {
	return format(u, "%r.%t")
}

// paths returns the path portion of the URL. e.g.
// for http://sub.example.com/path it will return
// []string{"/path"}
func paths(u *url.URL, f string) []string {
	return format(u, "%p")
}

// format is a little bit like a special sprintf for
// URLs; it will return a single formatted string
// based on the URL and the format string. e.g. for
// http://example.com/path and format string "%d%p"
// it will return example.com/path
func format(u *url.URL, f string) []string {
	out := &bytes.Buffer{}

	inFormat := false
	for _, r := range f {

		if r == '%' && !inFormat {
			inFormat = true
			continue
		}

		if !inFormat {
			out.WriteRune(r)
			continue
		}

		switch r {

		// a literal percent rune
		case '%':
			out.WriteRune('%')

		// the scheme; e.g. http
		case 's':
			out.WriteString(u.Scheme)

		// the userinfo; e.g. user:pass
		case 'u':
			if u.User != nil {
				out.WriteString(u.User.String())
			}

		// the domain; e.g. sub.example.com
		case 'd':
			out.WriteString(u.Hostname())

		// the port; e.g. 8080
		case 'P':
			out.WriteString(u.Port())

		// the subdomain; e.g. www
		case 'S':
			out.WriteString(extractFromDomain(u, "subdomain"))

		// the root; e.g. example
		case 'r':
			out.WriteString(extractFromDomain(u, "root"))

		// the tld; e.g. com
		case 't':
			out.WriteString(extractFromDomain(u, "tld"))

		// the path; e.g. /users
		case 'p':
			out.WriteString(u.EscapedPath())

		// the paths's file extension
		case 'e':
			paths := strings.Split(u.EscapedPath(), "/")
			if len(paths) > 1 {
				parts := strings.Split(paths[len(paths)-1], ".")
				if len(parts) > 1 {
					out.WriteString(parts[len(parts)-1])
				}
			} else {
				parts := strings.Split(u.EscapedPath(), ".")
				if len(parts) > 1 {
					out.WriteString(parts[len(parts)-1])
				}
			}

		// the query string; e.g. one=1&two=2
		case 'q':
			out.WriteString(u.RawQuery)

		// the fragment / hash value; e.g. section-1
		case 'f':
			out.WriteString(u.Fragment)

		// an @ if user info is specified
		case '@':
			if u.User != nil {
				out.WriteRune('@')
			}

		// a colon if a port is specified
		case ':':
			if u.Port() != "" {
				out.WriteRune(':')
			}

		// a question mark if there's a query string
		case '?':
			if u.RawQuery != "" {
				out.WriteRune('?')
			}

		// a hash if there is a fragment
		case '#':
			if u.Fragment != "" {
				out.WriteRune('#')
			}

		// the authority; e.g. user:pass@example.com:8080
		case 'a':
			out.WriteString(format(u, "%u%@%d%:%P")[0])

		// default to literal
		default:
			// output untouched
			out.WriteRune('%')
			out.WriteRune(r)
		}

		inFormat = false
	}

	return []string{out.String()}
}

func extractFromDomain(u *url.URL, selection string) string {

	// remove the port before parsing
	portRe := regexp.MustCompile(`(?m):\d+$`)

	domain := portRe.ReplaceAllString(u.Host, "")

	switch selection {
	case "subdomain":
		return extractor.GetSubdomain(domain)
	case "root":
		return extractor.GetDomain(domain)
	case "tld":
		return extractor.GetTld(domain)
	default:
		return ""
	}
}
