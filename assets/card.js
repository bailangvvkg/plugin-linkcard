// Linkcard: fills in what a card's block left out. Icons come from a favicon
// service; titles, descriptions and pictures can be fetched from a link
// detail service, if the site's settings allow it. Everything a card already
// says, nothing touches.
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

  function iconUrl(card) {
    var host = hostOf(card);
    if (!host) return "";
    switch (S.icons) {
      case "none":
        return "";
      case "google":
        return "https://www.google.com/s2/favicons?domain=" + encodeURIComponent(host) + "&sz=64";
      default:
        return "https://icons.duckduckgo.com/ip3/" + host + ".ico";
    }
  }

  function putIcon(card, src) {
    var body = card.querySelector(".kc-body");
    if (!body || body.querySelector(".kc-icon")) return;
    var img = document.createElement("img");
    img.className = "kc-icon";
    img.alt = "";
    img.loading = "lazy";
    img.src = src;
    body.insertBefore(img, body.firstChild);
    done(card, "icon");
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
    var src = iconUrl(card);
    if (src) putIcon(card, src);
  });

  // What is left needs asking a service, one card at a time, so the daily
  // limit of a free one lasts: the reader who opens a page pays for it once.
  var waiting = cards.filter(function (card) {
    if (S.fetch === false) return false;
    return needs(card, "title") || needs(card, "desc") || needs(card, "image");
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
    fetch(endpointOf(href))
      .then(function (r) {
        return r.ok ? r.json() : null;
      })
      .then(function (reply) {
        var meta = pick(reply);
        if (meta.title || meta.desc || meta.image || meta.icon) {
          writeCache(href, meta);
          apply(card, meta);
        }
      })
      .catch(function () {
        /* the card keeps what it was built with */
      })
      .then(fillNext);
  }

  fillNext();
})();
