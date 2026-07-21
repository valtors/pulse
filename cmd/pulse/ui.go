package main

const uiHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta http-equiv="Cache-Control" content="no-cache">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>pulse</title>
<style>
:root {
  --bg: #0a0c0f;
  --surface: #11151a;
  --border: #1a1d23;
  --border-hi: #2a2d33;
  --text: #e8ecf0;
  --text-mid: #8b95a0;
  --text-low: #4a5560;
  --accent: #22c55e;
  --urgent: #ef4444;
  --important: #f59e0b;
  --noise: #4a5560;
}
* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  background: var(--bg);
  color: var(--text);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
  font-size: 15px;
  line-height: 1.6;
  min-height: 100vh;
}
.app { max-width: 720px; margin: 0 auto; padding: 48px 20px 80px; }
.header { margin-bottom: 48px; }
.logo {
  font-size: 24px;
  font-weight: 800;
  letter-spacing: -0.5px;
  color: var(--text);
}
.tagline {
  color: var(--text-mid);
  font-size: 14px;
  margin-top: 4px;
}
.badges { display: flex; gap: 6px; flex-wrap: wrap; margin: 20px 0 32px; }
.badge {
  padding: 3px 10px;
  border-radius: 3px;
  font-size: 12px;
  background: var(--surface);
  border: 1px solid var(--border);
  color: var(--text-low);
  font-family: 'SF Mono', Monaco, monospace;
}
.badge.on {
  color: var(--accent);
  border-color: rgba(34, 197, 94, 0.2);
}
.section { margin-bottom: 32px; }
.section-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 2px;
  color: var(--text-low);
  margin-bottom: 12px;
}
.input-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}
input, textarea {
  flex: 1;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 12px 14px;
  color: var(--text);
  font-size: 14px;
  font-family: inherit;
  transition: border-color 0.15s;
}
input:focus, textarea:focus {
  outline: none;
  border-color: var(--border-hi);
}
button {
  background: var(--surface);
  border: 1px solid var(--border-hi);
  border-radius: 6px;
  padding: 10px 20px;
  color: var(--text);
  font-size: 14px;
  cursor: pointer;
  transition: background 0.15s;
  font-family: inherit;
  white-space: nowrap;
}
button:hover { background: var(--border-hi); }
button.primary {
  background: var(--accent);
  color: #000;
  border-color: var(--accent);
  font-weight: 600;
}
button.primary:hover { opacity: 0.9; }
.digest-box {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 20px;
  min-height: 80px;
  white-space: pre-wrap;
  font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-mid);
}
.digest-box .urgent { color: var(--urgent); font-weight: 600; }
.digest-box .important { color: var(--important); font-weight: 600; }
.digest-box .noise { color: var(--text-low); }
.ask-box {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 20px;
  min-height: 60px;
  white-space: pre-wrap;
  font-size: 14px;
  line-height: 1.7;
  color: var(--text);
  margin-top: 12px;
}
.thinking { color: var(--text-low); font-style: italic; }
.memory-item {
  padding: 10px 0;
  border-bottom: 1px solid var(--border);
}
.memory-item:last-child { border-bottom: none; }
.memory-key {
  font-weight: 600;
  color: var(--text);
  font-family: 'SF Mono', Monaco, monospace;
  font-size: 13px;
}
.memory-val {
  color: var(--text-mid);
  font-size: 13px;
  margin-top: 2px;
}
.memory-cat {
  font-size: 10px;
  text-transform: uppercase;
  color: var(--text-low);
  margin-left: 8px;
}
.connect-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-bottom: 12px;
}
@media (max-width: 600px) {
  .connect-grid { grid-template-columns: 1fr; }
  .app { padding: 32px 16px 60px; }
  .logo { font-size: 20px; }
}
.fade-in { animation: fadeIn 0.3s ease; }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
</style>
</head>
<body>
<div class="app">
  <div class="header">
    <div class="logo">pulse</div>
    <div class="tagline">connect everything. your ai does the rest.</div>
  </div>

  <div class="badges" id="badges"></div>

  <div class="section">
    <div class="section-label">connect a service</div>
    <div class="connect-grid">
      <input id="service" placeholder="github" />
      <input id="token" placeholder="token" type="password" />
    </div>
    <button onclick="connect()">connect</button>
  </div>

  <div class="section">
    <div class="section-label">digest</div>
    <button class="primary" onclick="digest()">what did i miss</button>
    <div class="digest-box" id="digest">click to get your filtered summary.</div>
  </div>

  <div class="section">
    <div class="section-label">ask</div>
    <div class="input-row">
      <input id="ask-input" placeholder="ask anything..." onkeydown="if(event.key==='Enter')ask()" autofocus />
      <button onclick="ask()">ask</button>
    </div>
    <div class="ask-box" id="ask-output"></div>
  </div>

  <div class="section">
    <div class="section-label">memory</div>
    <div class="input-row">
      <input id="mem-key" placeholder="key" />
      <input id="mem-val" placeholder="value" />
      <button onclick="remember()">store</button>
    </div>
    <div id="memory-list"></div>
  </div>
</div>

<script>
function status() {
  fetch('/api/status').then(r => r.json()).then(d => {
    const el = document.getElementById('badges');
    el.innerHTML = '';
    (d.connected || []).forEach(s => {
      el.innerHTML += '<span class="badge on">' + s + '</span>';
    });
    if (d.llm) el.innerHTML += '<span class="badge on">ai</span>';
    el.innerHTML += '<span class="badge">v' + (d.version || '?') + '</span>';
  });
}

function connect() {
  const service = document.getElementById('service').value.trim();
  const token = document.getElementById('token').value.trim();
  if (!service || !token) return;
  fetch('/api/connect', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({service,token})})
    .then(r => r.ok ? r.json() : Promise.reject(r.text()))
    .then(() => { status(); document.getElementById('service').value = ''; document.getElementById('token').value = ''; })
    .catch(e => { alert('connect failed'); });
}

function digest() {
  const el = document.getElementById('digest');
  el.innerHTML = '<span class="thinking">filtering...</span>';
  fetch('/api/digest')
    .then(r => r.json())
    .then(d => {
      el.innerHTML = '';
      if (d.ai_summary) {
        el.innerHTML += d.ai_summary + '\n\n---\n';
      }
      if (d.summary) {
        let html = d.summary
          .replace(/URGENT/g, '<span class="urgent">URGENT</span>')
          .replace(/NEEDS ATTENTION/g, '<span class="important">NEEDS ATTENTION</span>')
          .replace(/NOISE:/g, '<span class="noise">NOISE:</span>');
        el.innerHTML += html;
      }
      el.className = 'digest-box fade-in';
    })
    .catch(e => { el.innerHTML = 'error: ' + e; });
}

function ask() {
  const input = document.getElementById('ask-input').value.trim();
  if (!input) return;
  const el = document.getElementById('ask-output');
  el.innerHTML = '<span class="thinking">thinking...</span>';
  fetch('/api/ask', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({input})})
    .then(r => r.json())
    .then(d => {
      el.textContent = d.detail || d.summary || JSON.stringify(d, null, 2);
      el.className = 'ask-box fade-in';
    })
    .catch(e => { el.textContent = 'error: ' + e; });
}

function remember() {
  const key = document.getElementById('mem-key').value.trim();
  const val = document.getElementById('mem-val').value.trim();
  if (!key || !val) return;
  fetch('/api/ask', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({input:'remember ' + key + ' ' + val})})
    .then(() => {
      document.getElementById('mem-key').value = '';
      document.getElementById('mem-val').value = '';
      showMemory();
    });
}

function showMemory() {
  fetch('/api/memory').then(r => r.json()).then(d => {
    const el = document.getElementById('memory-list');
    if (!d || d.length === 0) {
      el.innerHTML = '<div class="memory-item" style="color:var(--text-low)">nothing remembered yet.</div>';
      return;
    }
    el.innerHTML = '';
    d.forEach(m => {
      el.innerHTML += '<div class="memory-item"><span class="memory-key">' + m.key + '</span><span class="memory-cat">' + m.category + '</span><div class="memory-val">' + m.value + '</div></div>';
    });
  });
}

status();
showMemory();
</script>
</body>
</html>`
