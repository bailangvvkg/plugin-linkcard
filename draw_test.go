package main

import (
	"strings"
	"testing"
)

// A card as Kite writes it: an unknown language renders plainly.
const plainCard = `<pre><code class="language-card">url: https://example.com/post
title: A post
</code></pre>`

// The shape a block chroma knows would take, so cards keep working if one
// of the names comes to chroma.
const chromaCard = `<pre class="chroma" data-lang="card"><code><span class="line"><span class="cl">https://example.com
</span></span></code></pre>`

func TestParams(t *testing.T) {
	for _, test := range []struct {
		in   string
		want map[string]string
	}{
		{"url: https://a.com\ntitle = \"A\"\n# comment\ntheme=grid\n", map[string]string{"url": "https://a.com", "title": "A", "size": "grid"}},
		{"https://a.com\n", map[string]string{"url": "https://a.com"}},
		{"unknown: x\nurl: https://a.com\n", map[string]string{"url": "https://a.com"}},
	} {
		got := params(test.in)
		for key, want := range test.want {
			if got[key] != want {
				t.Errorf("params(%q)[%q] = %q, want %q", test.in, key, got[key], want)
			}
		}
		if len(got) != len(test.want) {
			t.Errorf("params(%q) = %v, want exactly %v", test.in, got, test.want)
		}
	}
}

func TestDrawPlainShape(t *testing.T) {
	card, ok := draw(plainCard, "regular", "_blank")
	if !ok {
		t.Fatal("the plain shape was not drawn")
	}
	for _, want := range []string{
		`<a class="kite-card" data-size="regular" href="https://example.com/post"`,
		`<strong class="kc-title">A post</strong>`,
		`<span class="kc-host">example.com</span>`,
		` rel="noopener"`,
	} {
		if !strings.Contains(card, want) {
			t.Errorf("card is missing %q:\n%s", want, card)
		}
	}
}

func TestDrawChromaShape(t *testing.T) {
	card, ok := draw(chromaCard, "small", "_self")
	if !ok {
		t.Fatal("the chroma shape was not drawn")
	}
	if !strings.Contains(card, `data-size="small"`) {
		t.Errorf("card is not small:\n%s", card)
	}
	if !strings.Contains(card, `data-fill="title"`) {
		t.Errorf("a card without a title asks to have it filled:\n%s", card)
	}
	if got := strings.Contains(card, "desc"); got {
		t.Errorf("a small card asks for a description it cannot show:\n%s", card)
	}
}

func TestDrawRefuses(t *testing.T) {
	for _, bad := range []string{
		`<pre><code class="language-card">title: no address</code></pre>`,
		`<pre><code class="language-card">url: javascript:alert(1)</code></pre>`,
		`<pre><code class="language-card">url: not a url at all</code></pre>`,
	} {
		if _, ok := draw(bad, "regular", "_blank"); ok {
			t.Errorf("drew a card from %q", bad)
		}
	}
}

func TestDrawEscapes(t *testing.T) {
	card, ok := draw(`<pre><code class="language-card">url: https://a.com
title: &lt;b&gt;bold&lt;/b&gt;
</code></pre>`, "grid", "_blank")
	if !ok {
		t.Fatal("the card was not drawn")
	}
	if !strings.Contains(card, "&lt;b&gt;bold&lt;/b&gt;") {
		t.Errorf("the title was not escaped back into HTML:\n%s", card)
	}
	if strings.Contains(card, "<b>") {
		t.Errorf("raw markup got through:\n%s", card)
	}
}

func TestDrawRelativeAddress(t *testing.T) {
	card, ok := draw(`<pre><code class="language-card">url: /about/me/
</code></pre>`, "regular", "_blank")
	if !ok {
		t.Fatal("the card was not drawn")
	}
	if !strings.Contains(card, `href="/about/me/"`) {
		t.Errorf("the address was not kept:\n%s", card)
	}
	if strings.Contains(card, "kc-host") {
		t.Errorf("a card with no host shows none:\n%s", card)
	}
}

func TestDrawHaloTag(t *testing.T) {
	tag := `<hyperlink-card target="_blank" href="https://status.anye.xyz/status/docker" theme="regular" style="margin-top: 0.75em;"><a href="https://status.anye.xyz/status/docker" target="_blank">https://status.anye.xyz/status/docker</a></hyperlink-card>`
	card, ok := draw(tag, "regular", "_blank")
	if !ok {
		t.Fatal("the Halo tag was not drawn")
	}
	for _, want := range []string{
		`data-size="regular"`,
		`href="https://status.anye.xyz/status/docker"`,
		`<span class="kc-host">status.anye.xyz</span>`,
	} {
		if !strings.Contains(card, want) {
			t.Errorf("card is missing %q:\n%s", want, card)
		}
	}
	if strings.Contains(card, "hyperlink-card") {
		t.Errorf("the tag inside the card was kept:\n%s", card)
	}
}

func TestDrawHaloInlineTag(t *testing.T) {
	tag := `<hyperlink-inline-card href="https://www.kite.plus"></hyperlink-inline-card>`
	card, ok := draw(tag, "regular", "_blank")
	if !ok {
		t.Fatal("the Halo inline tag was not drawn")
	}
	if !strings.Contains(card, `data-size="small"`) {
		t.Errorf("an inline card is the small one:\n%s", card)
	}
}

func TestDrawHaloCustomFields(t *testing.T) {
	tag := `<hyperlink-card href="https://a.com" custom-title="A" custom-description="Ad" custom-image="i.png" custom-icon="c.ico"></hyperlink-card>`
	card, ok := draw(tag, "grid", "_blank")
	if !ok {
		t.Fatal("the Halo tag was not drawn")
	}
	for _, want := range []string{`>A</strong>`, `Ad`, `i.png`, `c.ico`} {
		if !strings.Contains(card, want) {
			t.Errorf("card is missing %q:\n%s", want, card)
		}
	}
}
