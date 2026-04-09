package renderer

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"
)

// ansiToHTML converts ANSI color codes to HTML <span> tags.
func ansiToHTML(s string) string {
	colorMap := map[string]string{
		"\033[0m":  "</span>",
		"\033[1m":  `<span style="font-weight:bold">`,
		"\033[31m": `<span style="color:red">`,
		"\033[32m": `<span style="color:green">`,
		"\033[33m": `<span style="color:#cc8800">`,
		"\033[35m": `<span style="color:magenta">`,
		"\033[36m": `<span style="color:teal">`,
		"\033[37m": `<span style="color:black">`,
	}

	// First HTML-escape the text (but preserve ANSI sequences).
	// We do this by splitting on ANSI codes, escaping each segment, then rejoining.
	type segment struct {
		text string
		ansi bool
	}
	var segments []segment
	rest := s
	for {
		idx := strings.Index(rest, "\033[")
		if idx < 0 {
			segments = append(segments, segment{rest, false})
			break
		}
		if idx > 0 {
			segments = append(segments, segment{rest[:idx], false})
		}
		// Find end of ANSI sequence.
		end := idx + 2
		for end < len(rest) && !((rest[end] >= 'A' && rest[end] <= 'Z') || (rest[end] >= 'a' && rest[end] <= 'z')) {
			end++
		}
		if end < len(rest) {
			end++
		}
		segments = append(segments, segment{rest[idx:end], true})
		rest = rest[end:]
	}

	var b strings.Builder
	for _, seg := range segments {
		if seg.ansi {
			if replacement, ok := colorMap[seg.text]; ok {
				b.WriteString(replacement)
			}
			// Unknown ANSI codes are dropped.
		} else {
			b.WriteString(html.EscapeString(seg.text))
		}
	}
	return b.String()
}

// StreamFrames writes ANSI frames to the response with chunked encoding (for terminals).
func StreamFrames(w http.ResponseWriter, frames []Frame, fast, instant bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		io.WriteString(w, frames[len(frames)-1].Content)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	for _, frame := range frames {
		io.WriteString(w, frame.Content)
		flusher.Flush()
		if !instant && frame.Delay > 0 {
			d := frame.Delay
			if fast {
				d /= 2
			}
			time.Sleep(d)
		}
	}
}

// StreamSSE writes frames as Server-Sent Events with HTML color markup (for browsers).
func StreamSSE(w http.ResponseWriter, frames []Frame, fast, instant bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	for _, frame := range frames {
		if frame.Tag == "engagement" {
			io.WriteString(w, "event: engagement\n")
			for _, line := range frame.Lines {
				colored := ansiToHTML(line)
				fmt.Fprintf(w, "data: %s\n", colored)
			}
			io.WriteString(w, "\n")
			flusher.Flush()

			if len(frame.Meta) > 0 {
				io.WriteString(w, "event: actions\n")
				for k, v := range frame.Meta {
					fmt.Fprintf(w, "data: %s=%s\n", k, v)
				}
				io.WriteString(w, "\n")
				flusher.Flush()
			}
			continue
		}

		if frame.Tag == "jackpot" {
			io.WriteString(w, "event: jackpot\n")
			for _, line := range frame.Lines {
				colored := ansiToHTML(line)
				fmt.Fprintf(w, "data: %s\n", colored)
			}
			io.WriteString(w, "\n")
			flusher.Flush()

			if !instant && frame.Delay > 0 {
				d := frame.Delay
				if fast {
					d /= 2
				}
				time.Sleep(d)
			}
			continue
		}

		for _, line := range frame.Lines {
			colored := ansiToHTML(line)
			fmt.Fprintf(w, "data: %s\n", colored)
		}
		io.WriteString(w, "\n")
		flusher.Flush()

		if !instant && frame.Delay > 0 {
			d := frame.Delay
			if fast {
				d /= 2
			}
			time.Sleep(d)
		}
	}

	io.WriteString(w, "event: done\ndata: \n\n")
	flusher.Flush()
}

// BrowserPage returns a minimal HTML page with SSE animation and a curl command builder.
// configForm is the HTML for game-specific form fields.
func BrowserPage(w http.ResponseWriter, title, path, configForm string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, browserHTML, title, configForm, path)
}

const browserHTML = `<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s — LUCK</title>
<style>
  body { margin: 8px; padding: 0; }
  pre, tt, code { font-family: 'Courier New', Courier, monospace; }
  pre { overflow-x: auto; }
  input, select { font-size: 16px; padding: 4px; box-sizing: border-box; }
  input[type=button] { padding: 8px 16px; }
  .field { margin: 6px 0; }
  .field b { display: inline-block; min-width: 70px; }
  #curl-wrap { overflow-x: auto; background: #fff; border: 1px solid #000;
               padding: 6px 8px; margin: 6px 0; font-family: 'Courier New', Courier, monospace; }
  #curl-wrap tt { white-space: nowrap; }
  @media (max-width: 600px) {
    pre { font-size: 12px; }
    .field b { display: block; min-width: 0; margin-bottom: 2px; }
    input[type=number], input[type=text], select { width: 100%%; max-width: 250px; }
  }
</style>
</head>
<body bgcolor="#c0c0c0">
<center>
<h1>LUCK</h1>
<p>
  <a href="/">Home</a> |
  <a href="/roulette">Roulette</a> |
  <a href="/slots">Slots</a> |
  <a href="/coinflip">Coin Flip</a> |
  <a href="/dice">Dice</a> |
  <a href="/double">Double</a>
</p>
<hr>
</center>
%s
<p>
  <input type="button" value="Go!" id="go" onclick="run()">
</p>
<pre id="out">Press Go! to play.</pre>
<div id="jp"></div>
<div id="eng"></div>
<hr>
<p><b>curl command:</b></p>
<div id="curl-wrap">
<tt id="curl">curl -N ...</tt>
</div>
<p><input type="button" value="Copy" onclick="copyCmd()"></p>
<p><font size="2"><i>LUCK Unleashes Chaotic Kindness</i></font></p>
<script>
var basePath = '%s';
function prefillForm() {
  var form = document.getElementById('cfg');
  if (!form) return;
  var params = new URLSearchParams(location.search);
  var inputs = form.querySelectorAll('input,select');
  for (var i = 0; i < inputs.length; i++) {
    var el = inputs[i];
    if (el.name && params.has(el.name)) {
      el.value = params.get(el.name);
    }
  }
}
function getFormQS() {
  var form = document.getElementById('cfg');
  if (!form) return '';
  var params = [];
  var inputs = form.querySelectorAll('input,select');
  for (var i = 0; i < inputs.length; i++) {
    var el = inputs[i];
    if (el.name && el.value && el.value !== el.getAttribute('data-default')) {
      params.push(encodeURIComponent(el.name) + '=' + encodeURIComponent(el.value));
    }
  }
  return params.length ? '?' + params.join('&') : '';
}
function buildSSEUrl() {
  var params = new URLSearchParams();
  var urlParams = new URLSearchParams(location.search);
  ['s','sc','h','u','cashout'].forEach(function(k) {
    if (urlParams.has(k)) params.set(k, urlParams.get(k));
  });
  var form = document.getElementById('cfg');
  if (form) {
    var inputs = form.querySelectorAll('input,select');
    for (var i = 0; i < inputs.length; i++) {
      var el = inputs[i];
      if (el.name && el.value && el.value !== el.getAttribute('data-default')) {
        params.set(el.name, el.value);
      } else if (el.name) {
        params.delete(el.name);
      }
    }
  }
  params.set('sse', '1');
  return basePath + '?' + params.toString();
}
function updateCurl() {
  var qs = getFormQS();
  var url = location.host + basePath + qs;
  if (qs) {
    document.getElementById('curl').textContent = 'curl -N "' + url + '"';
  } else {
    document.getElementById('curl').textContent = 'curl -N ' + url;
  }
}
function copyCmd() {
  var t = document.getElementById('curl').textContent;
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(t);
  } else {
    var ta = document.createElement('textarea');
    ta.value = t;
    document.body.appendChild(ta);
    ta.select();
    document.execCommand('copy');
    document.body.removeChild(ta);
  }
}
function toPath(url) {
  var i = url.indexOf('/');
  return i >= 0 ? url.substring(i) : '/' + url;
}
function run() {
  var btn = document.getElementById('go');
  btn.disabled = true;
  var out = document.getElementById('out');
  var eng = document.getElementById('eng');
  var jp = document.getElementById('jp');
  out.innerHTML = '';
  eng.innerHTML = '';
  jp.innerHTML = '';
  var src = new EventSource(buildSSEUrl());
  src.onmessage = function(e) {
    out.innerHTML = e.data;
  };
  src.addEventListener('jackpot', function(e) {
    jp.innerHTML = '<pre>' + e.data + '</pre>';
  });
  src.addEventListener('engagement', function(e) {
    eng.innerHTML = '<pre>' + e.data + '</pre>';
  });
  src.addEventListener('actions', function(e) {
    var lines = e.data.split('\n');
    var urls = {};
    for (var j = 0; j < lines.length; j++) {
      var eq = lines[j].indexOf('=');
      if (eq > 0) urls[lines[j].substring(0, eq)] = lines[j].substring(eq + 1);
    }
    var p = document.createElement('p');
    if (urls.nextURL) {
      var again = document.createElement('input');
      again.type = 'button'; again.value = 'Go Again!';
      again.onclick = function() { location.href = toPath(urls.nextURL); };
      p.appendChild(again); p.appendChild(document.createTextNode(' '));
    }
    if (urls.doubleURL) {
      var db = document.createElement('input');
      db.type = 'button'; db.value = 'Double or Nothing!';
      db.onclick = function() { location.href = toPath(urls.doubleURL); };
      p.appendChild(db); p.appendChild(document.createTextNode(' '));
    }
    if (urls.cashOutURL) {
      var co = document.createElement('input');
      co.type = 'button'; co.value = 'Cash Out!';
      co.onclick = function() { location.href = toPath(urls.cashOutURL); };
      p.appendChild(co);
    }
    eng.appendChild(p);
  });
  src.addEventListener('done', function() {
    src.close();
    btn.disabled = false;
  });
  src.onerror = function() {
    src.close();
    btn.disabled = false;
  };
}
document.addEventListener('DOMContentLoaded', function() {
  prefillForm();
  var form = document.getElementById('cfg');
  if (form) {
    var inputs = form.querySelectorAll('input,select');
    for (var i = 0; i < inputs.length; i++) {
      inputs[i].addEventListener('input', updateCurl);
      inputs[i].addEventListener('change', updateCurl);
    }
  }
  updateCurl();
  run();
});
</script>
</body>
</html>`
