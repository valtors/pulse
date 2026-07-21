package main

const uiHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta http-equiv="Cache-Control" content="no-cache">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>pulse</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
:root {
  --bg: #0a0a0a;
  --surface: #111;
  --surface-hi: #181818;
  --border: #1a1a1a;
  --border-hi: #2a2a2a;
  --text: #e5e5e5;
  --text-mid: #999;
  --text-low: #555;
  --accent: #22c55e;
  --urgent: #ef4444;
  --important: #f59e0b;
}
body {
  background: var(--bg);
  color: var(--text);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
  font-size: 15px;
  line-height: 1.6;
  min-height: 100vh;
}
.app { max-width: 680px; margin: 0 auto; padding: 56px 24px 80px; }
.header { margin-bottom: 40px; }
.logo {
  font-size: 28px;
  font-weight: 800;
  letter-spacing: -0.5px;
  color: var(--text);
}
.tagline {
  color: var(--text-mid);
  font-size: 14px;
  margin-top: 4px;
}
.badges { display: flex; gap: 8px; flex-wrap: wrap; margin: 16px 0 40px; }
.badge {
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 12px;
  background: var(--surface);
  border: 1px solid var(--border);
  color: var(--text-low);
}
.badge.on {
  color: var(--accent);
  border-color: rgba(34, 197, 94, 0.3);
  background: rgba(34, 197, 94, 0.05);
}
.section { margin-bottom: 36px; }
.section-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 2px;
  color: var(--text-low);
  margin-bottom: 14px;
}
.input-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}
input {
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
input:focus {
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
  transition: all 0.15s;
  font-family: inherit;
  white-space: nowrap;
}
button:hover { background: var(--surface-hi); border-color: var(--text-low); }
button.primary {
  background: var(--accent);
  color: #000;
  border-color: var(--accent);
  font-weight: 600;
}
button.primary:hover { opacity: 0.85; }
.digest-box {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 20px 24px;
  min-height: 60px;
  white-space: pre-wrap;
  font-family: 'SF Mono', Monaco, monospace;
  font-size: 13px;
  line-height: 1.8;
  color: var(--text-mid);
}
.digest-box .urgent { color: var(--urgent); font-weight: 600; }
.digest-box .important { color: var(--important); font-weight: 600; }
.digest-box .noise { color: var(--text-low); }
.digest-box .ok { color: var(--accent); }
.ask-box {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 20px 24px;
  min-height: 48px;
  white-space: pre-wrap;
  font-size: 14px;
  line-height: 1.7;
  color: var(--text);
  margin-top: 12px;
}
.thinking { color: var(--text-low); font-style: italic; }
.memory-item {
  padding: 12px 16px;
  border: 1px solid var(--border);
  border-radius: 6px;
  margin-bottom: 6px;
  background: var(--surface);
}
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
  overflow: hidden;
  text-overflow: ellipsis;
  max-height: 60px;
}
.memory-cat {
  font-size: 10px;
  text-transform: uppercase;
  color: var(--text-low);
  margin-left: 8px;
  padding: 1px 6px;
  border-radius: 3px;
  background: var(--bg);
}
.connect-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-bottom: 12px;
}
.empty-state {
  color: var(--text-low);
  font-size: 14px;
  padding: 20px;
  text-align: center;
  border: 1px dashed var(--border);
  border-radius: 8px;
}
.fade-in { animation: fadeIn 0.2s ease; }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
@media (max-width: 600px) {
  .connect-grid { grid-template-columns: 1fr; }
  .app { padding: 36px 16px 60px; }
  .logo { font-size: 24px; }
}
</style>
</head>
<body>
<div class="app">
  <div class="header">
    <div class="logo">pulse</div>
    <div class="tagline">your ai forgot everything. now it doesn't.</div>
  </div>

  <div class="badges" id="badges"></div>

  <div class="section">
    <div class="section-label">digest</div>
    <button class="primary" onclick="digest()">what did i miss</button>
    <div class="digest-box" id="digest" style="margin-top:12px">press the button.</div>
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
    <div class="section-label">connect</div>
    <div class="connect-grid">
      <input id="service" placeholder="github" />
      <input id="token" placeholder="token" type="password" />
    </div>
    <button onclick="connect()">connect</button>
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
    .then(r => r.ok ? r.json() : Promise.reject())
    .then(() => {
      status();
      document.getElementById('service').value = '';
      document.getElementById('token').value = '';
    })
    .catch(() => alert('connect failed'));
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
          .replace(/NOISE:/g, '<span class="noise">NOISE:</span>')
          .replace(/filtered/g, '<span class="ok">filtered</span>');
        el.innerHTML += html;
      }
      el.className = 'digest-box fade-in';
    })
    .catch(e => { el.innerHTML = 'error'; });
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
    .catch(e => { el.textContent = 'error'; });
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
      el.innerHTML = '<div class="empty-state">nothing remembered yet.</div>';
      return;
    }
    const visible = d.filter(m => m.category !== 'history' && m.category !== 'config');
    if (visible.length === 0) {
      el.innerHTML = '<div class="empty-state">nothing remembered yet.</div>';
      return;
    }
    el.innerHTML = '';
    visible.forEach(m => {
      let val = m.value;
      if (val.length > 120) val = val.substring(0, 120) + '...';
      el.innerHTML += '<div class="memory-item"><span class="memory-key">' + m.key + '</span><span class="memory-cat">' + m.category + '</span><div class="memory-val">' + val + '</div></div>';
    });
  });
}

status();
showMemory();
</script>
</body>
</html>`