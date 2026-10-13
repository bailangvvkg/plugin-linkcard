// Linkcard: fills in what a card's block left out. Icons come from a
// favicon service, or straight from the linked site, or the letter of its
// name; titles, descriptions and pictures can be fetched from a link
// detail service, if the site's settings allow it. Everything a card
// already says, nothing touches.
(function () {
  "use strict";

  var S = (window.KiteLinkCard && window.KiteLinkCard.settings) || {};
  var cards = Array.prototype.slice.call(document.querySelectorAll(".kite-card"));
  if (cards.length === 0) return;

  function needs(card, what) {
    return (" " + (card.getAttribute("data-need") || "") + " ").indexOf(" " + what + " ") >= 0;
  }

  function done(card, what) {
    var left = (" " + (card.getAttribute("data-need") || "") + " ").replace(" " + what + " ", " ").trim();
    if (left) {
      card.setAttribute("data-need", left);
    } else {
      card.removeAttribute("data-need");
    }
  }

  function hostOf(card) {
    try {
      return new URL(card.href).hostname;
    } catch (e) {
      return "";
    }
  }

  function iconUrls(card) {
    var host = hostOf(card);
    if (!host) return [];
    var origin = "https://" + host;
    try {
      origin = new URL(card.href).origin;
    } catch (e) {
      return [];
    }
    // The usual places a favicon sits, straight from the site the card
    // points at, through no one else: the last leg of every chain, for a
    // service that does not answer.
    var direct = [origin + "/favicon.ico", origin + "/apple-touch-icon.png", origin + "/favicon.png"];
    switch (S.icons) {
      case "none":
        return [];
      case "google":
        return ["https://www.google.com/s2/favicons?domain=" + encodeURIComponent(host) + "&sz=64"].concat(direct);
      case "direct":
        return direct;
      default:
        return ["https://icons.duckduckgo.com/ip3/" + host + ".ico"].concat(direct);
    }
  }

  function putIcon(card, src, more) {
    var body = card.querySelector(".kc-body");
    if (!body) return;
    var old = body.querySelector(".kc-icon");
    if (old && old.tagName === "IMG") return; // a real icon is in already
    if (old) old.remove(); // a letter was standing in for it
    var img = document.createElement("img");
    img.className = "kc-icon";
    img.alt = "";
    img.loading = "lazy";
    // Only a picture that truly came counts: one that fails to load leaves
    // the icon wanted, so a service's reply can still fill it.
    img.onload = function () {
      done(card, "icon");
    };
    img.onerror = function () {
      img.remove();
      if (more && more.length) {
        putIcon(card, more[0], more.slice(1));
      } else {
        letterIcon(card);
      }
    };
    img.src = src;
    body.insertBefore(img, body.firstChild);
  }

  // What every card can fall back to when nothing anywhere says what the
  // site's icon is: the first letter of its name.
  function letterIcon(card) {
    var body = card.querySelector(".kc-body");
    if (!body || body.querySelector(".kc-icon")) return;
    var host = hostOf(card);
    var span = document.createElement("span");
    span.className = "kc-icon kc-icon-letter";
    span.textContent = (host || "?").charAt(0).toUpperCase();
    body.insertBefore(span, body.firstChild);
    // Not done: a service's reply can still bring the real icon, which
    // takes the letter's place when it loads.
  }

  function putTitle(card, title) {
    if (!title) return;
    var el = card.querySelector(".kc-title");
    if (!el || el.getAttribute("data-fill") !== "title") return;
    el.textContent = title;
    el.removeAttribute("data-fill");
    done(card, "title");
  }

  function putDesc(card, desc) {
    if (!desc) return;
    if (card.getAttribute("data-size") === "small") return;
    if (card.querySelector(".kc-desc")) return;
    var host = card.querySelector(".kc-host");
    var span = document.createElement("span");
    span.className = "kc-desc";
    span.textContent = desc;
    if (host) {
      host.parentNode.insertBefore(span, host);
    } else {
      card.querySelector(".kc-text").appendChild(span);
    }
    done(card, "desc");
  }

  function putImage(card, src) {
    if (!src) return;
    var media = card.querySelector(".kc-media");
    if (!media) return;
    if (media.querySelector("img")) return;
    var img = document.createElement("img");
    img.alt = "";
    img.loading = "lazy";
    img.src = src;
    media.appendChild(img);
    media.removeAttribute("data-fill");
    done(card, "image");
  }

  function readCache(href) {
    if (!S.cache) return null;
    try {
      var raw = localStorage.getItem("kite-linkcard:v1:" + href);
      return raw ? JSON.parse(raw) : null;
    } catch (e) {
      return null;
    }
  }

  function writeCache(href, meta) {
    if (!S.cache) return;
    try {
      localStorage.setItem("kite-linkcard:v1:" + href, JSON.stringify(meta));
    } catch (e) {
      /* a full or absent store is nothing to stop for */
    }
  }

  function endpointOf(href) {
    if (S.api === "custom" && S.api_url) {
      return S.api_url.replace("{url}", encodeURIComponent(href));
    }
    return "https://api.microlink.io/?url=" + encodeURIComponent(href);
  }

  // The addresses to ask, best first: the one of the settings, then, so
  // that a service down for a moment leaves no card empty, Microlink
  // unless it was the one of the settings already.
  function endpointsOf(href) {
    var first = endpointOf(href);
    var ends = [first];
    if (first.indexOf("microlink.io") < 0) {
      ends.push("https://api.microlink.io/?url=" + encodeURIComponent(href));
    }
    return ends;
  }

  // A reply may be Microlink's shape, {status, data: {...}}, or plainly the
  // data itself, as a custom endpoint answers; pictures may be strings or
  // {url: ...} objects.
  function pick(reply) {
    var d = reply && reply.data && typeof reply.data === "object" && !reply.data.url ? reply.data : reply || {};
    function addr(v) {
      if (!v) return "";
      if (typeof v === "string") return v;
      if (v.url) return v.url;
      return "";
    }
    return {
      title: d.title || "",
      desc: d.description || d.desc || "",
      image: addr(d.image) || addr(d.screenshot),
      icon: addr(d.logo) || addr(d.icon || d.publisher)
    };
  }

  function apply(card, meta) {
    putTitle(card, meta.title);
    putDesc(card, meta.desc);
    putImage(card, meta.image);
    if (needs(card, "icon") && meta.icon) putIcon(card, meta.icon);
  }

  // Icons need no service call: point the picture at a favicon and be done.
  cards.forEach(function (card) {
    if (!needs(card, "icon")) return;
    var urls = iconUrls(card);
    if (urls.length) putIcon(card, urls[0], urls.slice(1));
  });

  // What is left needs asking a service, one card at a time, so the daily
  // limit of a free one lasts: the reader who opens a page pays for it once.
  var waiting = cards.filter(function (card) {
    if (S.fetch === false) return false;
    return needs(card, "title") || needs(card, "desc") || needs(card, "image") || needs(card, "icon");
  });

  function fillNext() {
    var card = waiting.shift();
    if (!card) return;
    var href = card.href; // absolute, whatever the card was written with
    var meta = readCache(href);
    if (meta) {
      apply(card, meta);
      fillNext();
      return;
    }
    ask(card, href, endpointsOf(href), 0);
  }

  // Asks each address in turn until one brings something of the card.
  function ask(card, href, ends, i) {
    if (i >= ends.length) {
      fillNext();
      return;
    }
    fetch(ends[i])
      .then(function (r) {
        return r.ok ? r.json() : null;
      })
      .then(function (reply) {
        var meta = pick(reply);
        if (meta.title || meta.desc || meta.image || meta.icon) {
          writeCache(href, meta);
          apply(card, meta);
          fillNext();
          return;
        }
        ask(card, href, ends, i + 1);
      })
      .catch(function () {
        /* this address said nothing: the next one is asked */
        ask(card, href, ends, i + 1);
      });
  }

  fillNext();

  // Whatever is still without an icon once every attempt had its turn —
  // a service that does not answer, a site with no favicon anywhere —
  // gets the letter of its name, so no card looks unfinished.
  setTimeout(function () {
    cards.forEach(function (card) {
      if (needs(card, "icon")) letterIcon(card);
    });
  }, 1200);
})();
