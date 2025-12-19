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
.session-nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  margin-bottom: 16px;
}
.session-nav a {
  color: #2563eb;
  text-decoration: none;
  font-size: 14px;
}
.session-nav a:hover {
  text-decoration: underline;
}
.session-nav .nav-next {
  margin-left: auto;
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
  background: #f8f8f8;
  border: 1px solid #e0e0e0;
  border-left: 3px solid #999;
  border-radius: 6px;
  color: #555;
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
.plan-approved {
  background: #f5f5f5;
  border: 1px solid #e0e0e0;
  color: #333;
}
.question-block {
  border-left: 3px solid #d0d0d0;
  padding-left: 12px;
}
.question-item {
  margin: 12px 0;
  padding: 12px;
  background: #fafafa;
  border-radius: 8px;
}
.question-header {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  color: #888;
  margin-bottom: 4px;
}
.question-text {
  font-size: 14px;
  color: #333;
  margin-bottom: 8px;
}
.question-options {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.question-option {
  display: flex;
  flex-direction: column;
  padding: 8px 12px;
  background: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
}
.option-label {
  font-weight: 500;
  color: #333;
  font-size: 13px;
}
.option-desc {
  font-size: 12px;
  color: #666;
  margin-top: 2px;
}
.question-result {
  margin: 8px 0;
  padding: 12px;
  background: #f8f8f8;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
}
.question-result-header {
  font-size: 12px;
  font-weight: 600;
  color: #555;
  margin-bottom: 8px;
}
.answer-item {
  display: flex;
  flex-direction: column;
  margin-bottom: 6px;
  padding-bottom: 6px;
  border-bottom: 1px solid #eee;
}
.answer-item:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}
.answer-question {
  font-size: 12px;
  color: #666;
}
.answer-value {
  font-size: 13px;
  color: #333;
  font-weight: 500;
}
.subagent-block {
  background: #f8f8f8;
  border: 1px solid #e8e8e8;
  border-left: 3px solid #999;
  border-radius: 8px;
  padding: 12px;
  margin: 4px 0;
}
.subagent-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.subagent-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: #666;
  color: white;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}
.subagent-badge svg {
  width: 12px;
  height: 12px;
  color: white;
}
.subagent-note {
  font-size: 11px;
  color: #888;
  font-style: italic;
}
.subagent-type {
  font-size: 13px;
  color: #333;
  font-weight: 500;
  margin-bottom: 8px;
}
.subagent-block .collapsible {
  margin: 0;
}
.todo-list {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.todo-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  background: #fafafa;
  border-radius: 6px;
  font-size: 13px;
}
.todo-status {
  font-size: 14px;
  width: 16px;
  text-align: center;
}
.todo-item.completed .todo-status { color: #22c55e; }
.todo-item.in-progress .todo-status { color: #3b82f6; }
.todo-item.pending .todo-status { color: #9ca3af; }
.todo-item.completed .todo-content { color: #666; text-decoration: line-through; }
.todo-item.in-progress .todo-content { color: #333; font-weight: 500; }
.todo-item.pending .todo-content { color: #555; }
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
NAV_PLACEHOLDER
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
