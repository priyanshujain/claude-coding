package template

const HTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>TITLE_PLACEHOLDER</title>
<link href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism.min.css" rel="stylesheet" />
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #fff;
  color: #1a1a1a;
  line-height: 1.7;
  font-size: 15px;
}
.container { max-width: 720px; margin: 0 auto; padding: 40px 20px; }
.header {
  text-align: center;
  padding-bottom: 32px;
  margin-bottom: 32px;
  border-bottom: 1px solid #eee;
}
.header h1 {
  font-size: 1.4rem;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 12px;
  line-height: 1.4;
}
.header .meta {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 0.875rem;
  color: #666;
}
.header .avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #e91e63;
  color: white;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 600;
}
.message {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  align-items: flex-start;
}
.message .avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}
.message.user .avatar {
  background: #e91e63;
  color: white;
}
.message.assistant .avatar {
  background: #f5f5f5;
  border: 1px solid #e0e0e0;
  color: #666;
}
.message.assistant .avatar svg {
  width: 18px;
  height: 18px;
}
.message-content {
  flex: 1;
  min-width: 0;
}
.message.user .message-content {
  background: #f8f9fa;
  padding: 14px 16px;
  border-radius: 12px;
}
.text-block {
  margin-bottom: 8px;
  white-space: pre-wrap;
  word-wrap: break-word;
}
.text-block:last-child { margin-bottom: 0; }
.collapsible {
  margin: 2px 0 12px 0;
}
.collapsible-header {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0;
  cursor: pointer;
  font-size: 14px;
  color: #666;
  user-select: none;
}
.collapsible-header:hover { color: #333; }
.collapsible-header .chevron {
  transition: transform 0.15s;
  font-size: 10px;
}
.collapsible.open .chevron { transform: rotate(90deg); }
.collapsible-content {
  display: none;
  margin-top: 8px;
  padding: 12px 16px;
  background: #fafafa;
  border-radius: 8px;
  font-size: 13px;
  color: #555;
  max-height: 300px;
  overflow: auto;
  white-space: pre-wrap;
}
.collapsible.open .collapsible-content { display: block; }
.tool-block {
  margin: 4px 0;
}
.tool-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: #f5f5f5;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  font-size: 13px;
  color: #555;
}
.tool-pill svg {
  width: 14px;
  height: 14px;
  color: #888;
}
.slash-command {
  display: inline-block;
  padding: 4px 10px;
  background: #e8e8e8;
  border-radius: 4px;
  font-size: 13px;
  font-family: monaco, ui-monospace, 'SF Mono', monospace;
  color: #555;
}
.command-block .tool-pill {
  background: #f0f0f0;
  border-color: #ddd;
  font-size: 12px;
  padding: 4px 10px;
  color: #666;
}
.local-output {
  padding: 6px 10px;
  background: #f8f8f8;
  border-left: 3px solid #ddd;
  font-size: 12px;
  color: #666;
  margin: 4px 0;
  font-family: monaco, ui-monospace, 'SF Mono', monospace;
}
.tool-info {
  margin-top: 6px;
  padding: 10px 12px;
  background: #fafafa;
  border-radius: 8px;
  font-size: 13px;
  color: #666;
}
.tool-info a {
  color: #2563eb;
  word-break: break-all;
}
.tool-result {
  margin-top: 8px;
}
.tool-result pre {
  margin: 0;
  background: #f5f5f5;
  color: #333;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  font-size: 13px;
  font-family: monaco, ui-monospace, 'SF Mono', monospace;
  max-height: 200px;
  overflow: auto;
}
.tool-result-inline {
  margin: 8px 0;
  padding: 8px 12px;
  background: #f5f5f5;
  border-radius: 6px;
  font-size: 13px;
  color: #666;
}
.tool-result-files {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin: 8px 0;
}
.tool-result-files .file-path {
  display: inline-block;
  padding: 4px 8px;
  background: #f5f5f5;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  font-size: 12px;
  font-family: monaco, ui-monospace, 'SF Mono', monospace;
  color: #555;
  cursor: default;
  width: fit-content;
}
.tool-result-files .file-path:hover {
  background: #eee;
}
.search-result {
  margin: 4px 0 20px 0;
  padding-left: 16px;
  border-left: 2px solid #e0e0e0;
}
.search-result-count {
  font-size: 13px;
  color: #666;
}
.search-result-list {
  margin-top: 4px;
}
.search-result-item {
  font-size: 12px;
  font-family: monaco, ui-monospace, 'SF Mono', monospace;
  color: #555;
  padding: 1px 0;
}
.diff-block {
  margin-top: 8px;
  border-radius: 8px;
  overflow: hidden;
  font-family: monaco, ui-monospace, 'SF Mono', monospace;
  font-size: 12px;
  border: 1px solid #d1d5da;
}
.diff-line {
  padding: 2px 10px;
  white-space: pre-wrap;
  word-wrap: break-word;
}
.diff-removed {
  background: #ffebe9;
  color: #82071e;
}
.diff-added {
  background: #e6ffec;
  color: #116329;
}
.bash-command {
  margin-top: 8px;
  padding: 8px 12px;
  background: #f5f5f5;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  overflow-x: auto;
}
.bash-command code {
  background: none;
  color: #333;
  padding: 0;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}
.tool-result-error {
  margin: 8px 0;
  padding: 8px 12px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  color: #dc2626;
  font-size: 12px;
  white-space: pre-wrap;
}
code {
  background: #f5f5f5;
  color: #333;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monaco, ui-monospace, 'SF Mono', monospace;
  font-size: 0.9em;
}
pre {
  background: #f5f5f5;
  color: #333;
  padding: 14px;
  border-radius: 8px;
  overflow-x: auto;
  font-family: monaco, ui-monospace, 'SF Mono', monospace;
  font-size: 13px;
  margin: 10px 0;
  line-height: 1.5;
  border: 1px solid #e0e0e0;
}
pre code { background: none; color: inherit; padding: 0; }
ul, ol { margin: 10px 0; padding-left: 20px; }
li { margin-bottom: 4px; line-height: 1.5; }
h2 { font-size: 1.2rem; font-weight: 600; margin: 16px 0 10px; color: #1a1a1a; }
h3 { font-size: 1.05rem; font-weight: 600; margin: 14px 0 8px; color: #1a1a1a; }
h4 { font-size: 1rem; font-weight: 600; margin: 12px 0 6px; color: #333; }
a { color: #2563eb; text-decoration: none; }
a:hover { text-decoration: underline; }
strong { font-weight: 600; }
</style>
</head>
<body>
<div class="container">
<div class="header">
<h1>TITLE_PLACEHOLDER</h1>
<div class="meta">
<span class="avatar">INITIALS_PLACEHOLDER</span>
<span>USERNAME_PLACEHOLDER</span>
</div>
</div>
MESSAGES_PLACEHOLDER
</div>
<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/prism.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-go.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-python.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-javascript.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-typescript.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-bash.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-json.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-yaml.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-markdown.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-rust.min.js"></script>
<script>
document.querySelectorAll('.collapsible-header').forEach(h => {
  h.addEventListener('click', () => h.closest('.collapsible').classList.toggle('open'));
});
Prism.highlightAll();
</script>
</body>
</html>`

const ClaudeIcon = `<img src="https://upload.wikimedia.org/wikipedia/commons/thumb/b/b0/Claude_AI_symbol.svg/960px-Claude_AI_symbol.svg.png" alt="Claude" style="width:20px;height:20px;">`
