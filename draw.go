// Draw turns the card code blocks of a page into link cards.
package main

import (
	"html"
	"net/url"
	"regexp"
	"strings"
)

// cards matches a fenced code block whose language is card. Kite writes
// one of two shapes: a language chroma does not know, as card is, renders
// plainly as <pre><code class="language-card">...</code></pre>, while one
// chroma knows would carry data-lang="card". The content is escaped either
// way, so it holds no tags of its own.
var cards = regexp.MustCompile(`(?s)<pre><code class="language-(?:card|linkcard|link-card)">.*?</code></pre>|<pre\b[^>]*\bdata-lang="(?:card|linkcard|link-card)"[^>]*>.*?</pre>`)

// hyperlinks matches the tags the Halo editor wrote, as
// <hyperlink-card href="..." theme="..." custom-title="...">...</hyperlink-card>,
// so a post moved from Halo keeps its cards. What the tag wraps, a plain
// link that stands in for the card, is dropped.
var hyperlinks = regexp.MustCompile(`(?is)<hyperlink-(?:inline-)?card\b[^>]*>.*?</hyperlink-(?:inline-)?card>`)

// articles matches one article region, where a page's writing is, so a
// link in a menu or a footer stays a link. The article tag is what themes
// wrap a single piece of writing in.
var articles = regexp.MustCompile(`(?is)<article\b.*?</article>`)

// anchors matches a link and what it says, so one whose text is just its
// own address can stand for a card.
var anchors = regexp.MustCompile(`(?i)<a href="(https?://[^"]*)"[^>]*>([^<]*)</a>`)

// bareCards turns the links of a page's writing into cards, when the
// setting asks for it: a link that says its own address, as an autolink
// does, is a card's one-line form.
func bareCards(pageHTML, size, target string) string {
	return articles.ReplaceAllStringFunc(pageHTML, func(region string) string {
		return anchors.ReplaceAllStringFunc(region, func(link string) string {
			m := anchors.FindStringSubmatch(link)
			if !saysAddress(m[2], m[1]) {
				return link
			}
			if card, ok := drawCard(map[string]string{"url": m[1]}, size, target); ok {
				return card
			}
			return link
		})
	})
}

// saysAddress tells whether a link's text is the address itself, with or
// without the scheme, the www, or the last slash.
func saysAddress(text, href string) bool {
	t := strings.TrimSpace(html.UnescapeString(text))
	if t == "" || len(t) > len(href)+7 {
		return false
	}
	h := html.UnescapeString(href)
	plain := strings.TrimPrefix(strings.TrimPrefix(h, "https://"), "http://")
	plain = strings.TrimPrefix(plain, "www.")
	t, h, plain = strings.TrimSuffix(t, "/"), strings.TrimSuffix(h, "/"), strings.TrimSuffix(plain, "/")
	return t == h || t == plain
}

// attrs reads the attributes a hyperlink card tag gives that matter here:
// its address, its size, and the parts its author wrote by hand.
var attrs = regexp.MustCompile(`(?is)(href|target|theme|custom-title|custom-description|custom-image|custom-icon)\s*=\s*"([^"]*)"`)

var attrNames = map[string]string{"href": "url", "target": "target", "theme": "size", "custom-title": "title", "custom-description": "desc", "custom-image": "image", "custom-icon": "icon"}

// tags strips the markup chroma adds, and params takes the keys and values
// from what a card's block says.
var (
	tags  = regexp.MustCompile(`<[^>]*>`)
	param = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_-]*)\s*[:=]\s*(.*)$`)
	sizes = []string{"small", "regular", "grid"}
	named = map[string]string{"url": "url", "link": "url", "href": "url", "title": "title", "name": "title", "desc": "desc", "description": "desc", "summary": "desc", "image": "image", "img": "image", "cover": "image", "icon": "icon", "favicon": "icon", "logo": "icon", "size": "size", "theme": "size", "target": "target"}
)

// blockText reads what a card's code block says, as the one who wrote it
// typed it: the tags around and inside it are gone and what they escaped is
// back. Both of the shapes a block takes come to the same thing this way.
func blockText(pre string) string {
	return html.UnescapeString(tags.ReplaceAllString(pre, ""))
}

// params reads a card's block: a key and value to a line, as title: ..., or
// key=value, and a line that is only an address is the address. Quoted values
// lose their quotes; unknown keys are left, so new ones can come.
func params(text string) map[string]string {
	m := map[string]string{}
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if v := param.FindStringSubmatch(line); v != nil {
			if key, ok := named[strings.ToLower(v[1])]; ok {
				m[key] = unquote(v[2])
				continue
			}
		}
		if _, have := m["url"]; !have && (strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") || strings.HasPrefix(line, "/")) {
			m["url"] = line
		}
	}
	return m
}

// unquote takes the quotes off a value that has a pair of them.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if n := len(s); n >= 2 {
		if (s[0] == '"' && s[n-1] == '"') || (s[0] == '\'' && s[n-1] == '\'') {
			return s[1 : n-1]
		}
	}
	return s
}

func knownSize(s string) bool {
	for _, size := range sizes {
		if size == s {
			return true
		}
	}
	return false
}

// draw writes the card a code block, or a hyperlink card tag moved from
// Halo, stands for. What it leaves out, the card asks for with data-need,
// and card.js fills it in.
func draw(tag, defSize, defTarget string) (string, bool) {
	p, ok := paramsOf(tag)
	if !ok {
		return "", false
	}
	return drawCard(p, defSize, defTarget)
}

// paramsOf reads what a card's markup says: a code block holds keys and
// values as text, a hyperlink tag as attributes.
func paramsOf(tag string) (map[string]string, bool) {
	if strings.HasPrefix(strings.ToLower(tag), "<hyperlink") {
		m := map[string]string{}
		for _, a := range attrs.FindAllStringSubmatch(tag, -1) {
			if key, ok := attrNames[strings.ToLower(a[1])]; ok {
				m[key] = a[2]
			}
		}
		// Halo's inline card is the small one.
		if strings.Contains(tag, "<hyperlink-inline-card") {
			if _, ok := m["size"]; !ok {
				m["size"] = "small"
			}
		}
		return m, m["url"] != ""
	}
	m := params(blockText(tag))
	return m, m["url"] != ""
}

// drawCard writes the card what a card's markup says stands for.
func drawCard(p map[string]string, defSize, defTarget string) (string, bool) {
	address := strings.TrimSpace(p["url"])
	if address == "" {
		return "", false
	}
	u, err := url.Parse(address)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https" && address[0] != '/') || (u.Hostname() == "" && address[0] != '/') {
		return "", false
	}

	host := u.Hostname()
	fallback := host
	if host == "" {
		fallback = strings.Trim(address, "/")
		if i := strings.LastIndex(fallback, "/"); i >= 0 {
			fallback = fallback[i+1:]
		}
	}

	size := defSize
	if !knownSize(size) {
		size = "regular"
	}
	if s := p["size"]; knownSize(s) {
		size = s
	}
	target := defTarget
	switch p["target"] {
	case "_self":
		target = "_self"
	case "_blank":
		target = "_blank"
	}

	var needs []string
	if p["icon"] == "" {
		needs = append(needs, "icon")
	}
	if p["title"] == "" {
		needs = append(needs, "title")
	}
	if p["desc"] == "" && size != "small" {
		needs = append(needs, "desc")
	}
	if size != "small" && p["image"] == "" {
		needs = append(needs, "image")
	}

	esc := html.EscapeString
	var b strings.Builder
	b.WriteString(`<a class="kite-card" data-size="` + size + `" href="` + esc(address) + `"`)
	if target != "" {
		b.WriteString(` target="` + target + `"`)
		if target == "_blank" {
			b.WriteString(` rel="noopener"`)
		}
	}
	if len(needs) > 0 {
		b.WriteString(` data-need="` + strings.Join(needs, " ") + `"`)
	}
	b.WriteString(`>`)

	// The picture comes before the text, so a grid card shows it on top by
	// the document's order and a regular one moves it to the side in CSS.
	if size != "small" {
		if image := p["image"]; image != "" {
			b.WriteString(`<span class="kc-media"><img src="` + esc(image) + `" alt="" loading="lazy"></span>`)
		} else {
			b.WriteString(`<span class="kc-media" data-fill="image"></span>`)
		}
	}

	b.WriteString(`<span class="kc-body">`)
	if icon := p["icon"]; icon != "" {
		b.WriteString(`<img class="kc-icon" src="` + esc(icon) + `" alt="" width="32" height="32" loading="lazy">`)
	}
	b.WriteString(`<span class="kc-text">`)
	title, fillTitle := p["title"], false
	if title == "" {
		title, fillTitle = fallback, true
	}
	b.WriteString(`<strong class="kc-title"`)
	if fillTitle {
		b.WriteString(` data-fill="title"`)
	}
	b.WriteString(`>` + esc(title) + `</strong>`)
	if desc := p["desc"]; desc != "" {
		b.WriteString(`<span class="kc-desc">` + esc(desc) + `</span>`)
	}
	if host != "" {
		b.WriteString(`<span class="kc-host">` + esc(host) + `</span>`)
	}
	b.WriteString(`</span></span></a>`)
	return b.String(), true
}
