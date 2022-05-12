package system

import (
	"fmt"
	"strings"
)

func (s *System) renderedHomePage(onlineNodes []string) ([]byte, string, error) {
	html := renderHeader() + s.renderOnlineNodes(onlineNodes) + renderFooter()
	return []byte(html), htmlContent, nil
}

func (s *System) renderOnlineNodes(nodes []string) string {
	nodesHtml := "<ul>"
	nodesHtml += fmt.Sprintf(`<li><strong><a href="http://localhost:%d/%s/index.html">%s</a></strong></li>`,
		s.cfg.LocalWebServerPort,
		s.cfg.Username,
		s.cfg.Username,
	)
	for _, node := range nodes {
		nodesHtml += fmt.Sprintf(`<li><strong><a href="http://localhost:%d/%s/index.html">%s</a></strong></li>`,
			s.cfg.LocalWebServerPort,
			node,
			node,
		)
	}
	nodesHtml += "</ul>"
	return nodesHtml
}

func renderHeader() string {
	return `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="user-scalable=0, width=device-width, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
		<title>Localhosts</title>
	<head>
	<body>
		<p><strong>Online Users</strong></p>
	`
}

func renderFooter() string {
	return `
	</body>
	</html>
	`
}

func inferContentType(path string) string {
	switch {
	case strings.HasSuffix(path, ".html"):
		return htmlContent
	case strings.HasSuffix(path, ".css"):
		return cssContent
	case strings.HasSuffix(path, ".js"):
		return jsContent
	case strings.HasSuffix(path, ".png"):
		return pngContent
	default:
		return ""
	}
}
