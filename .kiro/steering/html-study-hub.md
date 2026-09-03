---
inclusion: manual
---

# HTML Study Hub — Format & Design System (self-contained)

Steering này chứa **toàn bộ** CSS, JS và template HTML cần thiết để dựng "study hub" dạng
web deep-dive (dark theme, sidebar, flashcard toggle, sơ đồ vẽ bằng CSS) từ bộ tài liệu markdown
đã có. **Tự chứa** — không phụ thuộc file bên ngoài, dùng được ở bất kỳ project nào.

> Dùng steering này khi đã có các file `.md` (notes/flashcards/plan) và muốn sinh thêm lớp HTML,
> hoặc khi steering `study-material-generator` cần phần HTML. Kích hoạt: `#html-study-hub`.

---

## Quy trình dựng HTML hub

1. Xác định danh sách "trang" cần tạo (thường mỗi file notes/tuần = 1 trang `weekN.html`,
   cộng `index.html` làm landing).
2. Tạo `html/assets/style.css` — **copy nguyên văn** khối CSS ở mục *STYLE.CSS* bên dưới.
3. Tạo `html/assets/app.js` — **copy nguyên văn** khối JS ở mục *APP.JS* bên dưới.
4. **Rebrand**: đổi biến `--accent` (và tùy chọn `--accent-2`, `--accent-3`) trong `:root`
   cho hợp chủ đề (xem *REBRAND*).
5. Tạo `html/index.html` từ template *INDEX TEMPLATE*.
6. Với mỗi tuần/chủ đề: tạo `html/weekN.html` từ template *PAGE TEMPLATE*, chuyển nội dung
   markdown sang các component (block, table, diagram, flashcard).
7. **Verify**: số `<section>` = `</section>`; mọi `id` khớp `href="#..."` trong sidebar;
   nav "All Weeks" xuất hiện & đồng bộ trên mọi trang; đã escape `&amp; &lt; &gt;`.

### Chuyển markdown → component HTML

| Markdown | Component HTML |
|---|---|
| `## N. Tiêu đề` | `<section id="..."><h2 class="section-title"><span class="num">N</span> ...</h2>` |
| đoạn "vì sao/cơ chế" | `<div class="block why">` / `.how` / `.purpose` |
| "bẫy/lưu ý" | `<div class="block pitfall">` (đỏ) |
| "best practice/senior" | `<div class="block best">` (cam) |
| bảng markdown | `<table>` (đã style sẵn) |
| so sánh 2 phía | `<div class="grid-2"><div class="col">...` hoặc `.proscons` |
| tóm tắt đầu bài | `<div class="tldr">` |
| sơ đồ ASCII | ưu tiên vẽ lại bằng *DIAGRAM KIT* (node + arrow) |
| `### Q:` / `**A:**` | `<div class="flashcard">` (xem *FLASHCARD*) |
| code block | `<pre><code>` + tô màu bằng `.tok-*` span |

---

## REBRAND

Chỉ cần sửa vài biến trong `:root` của `style.css` để đổi màu thương hiệu theo chủ đề:

```
Go        --accent: #00add8  (cyan)
Rust      --accent: #dea584  (cam đất)
Python    --accent: #4b8bbe  (xanh dương)   --accent-2: #ffd43b (vàng)
Java      --accent: #f89820  (cam)          --accent-2: #5382a1 (xanh)
Node/JS   --accent: #83cd29  (xanh lá)      --accent-2: #f0db4f (vàng)
Kubernetes--accent: #326ce5  (xanh)
```

Logo/tiêu đề trong `.brand .logo` và `<title>` đổi theo tên chủ đề. Phần còn lại giữ nguyên.

---

## STYLE.CSS  (ghi vào `html/assets/style.css`)

```css
/* ===== Study Hub - Shared Styles ===== */
:root {
  --bg: #0d1117;
  --bg-alt: #161b22;
  --bg-card: #1c2128;
  --border: #30363d;
  --text: #e6edf3;
  --text-dim: #9198a1;
  --accent: #00add8;      /* ĐỔI theo chủ đề */
  --accent-2: #7ee787;    /* green */
  --accent-3: #ffa657;    /* orange */
  --danger: #ff7b72;      /* red */
  --purple: #d2a8ff;
  --code-bg: #0b0f14;
  --radius: 10px;
  --shadow: 0 4px 20px rgba(0,0,0,0.4);
}

* { box-sizing: border-box; margin: 0; padding: 0; }

html { scroll-behavior: smooth; }

body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  background: var(--bg);
  color: var(--text);
  line-height: 1.65;
  font-size: 16px;
}

/* ===== Layout ===== */
.layout { display: flex; min-height: 100vh; }

.sidebar {
  width: 280px;
  background: var(--bg-alt);
  border-right: 1px solid var(--border);
  padding: 24px 0;
  position: sticky;
  top: 0;
  height: 100vh;
  overflow-y: auto;
  flex-shrink: 0;
}

.sidebar h2 {
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: var(--text-dim);
  padding: 0 24px;
  margin: 20px 0 8px;
}

.sidebar .brand {
  padding: 0 24px 16px;
  border-bottom: 1px solid var(--border);
  margin-bottom: 12px;
}
.sidebar .brand a { text-decoration: none; }
.sidebar .brand .logo {
  font-size: 20px; font-weight: 700; color: var(--accent);
  display: flex; align-items: center; gap: 8px;
}
.sidebar .brand .sub { font-size: 12px; color: var(--text-dim); margin-top: 4px; }

.sidebar nav a {
  display: block;
  padding: 7px 24px;
  color: var(--text-dim);
  text-decoration: none;
  font-size: 14px;
  border-left: 3px solid transparent;
  transition: all .15s;
}
.sidebar nav a:hover { color: var(--text); background: var(--bg-card); }
.sidebar nav a.active { color: var(--accent); border-left-color: var(--accent); background: var(--bg-card); }

.content {
  flex: 1;
  padding: 48px 56px 120px;
  max-width: 960px;
  margin: 0 auto;
  width: 100%;
}

/* ===== Hero ===== */
.hero {
  background: linear-gradient(135deg, rgba(0,173,216,0.15), rgba(126,231,135,0.08));
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 32px 36px;
  margin-bottom: 40px;
}
.hero .week-badge {
  display: inline-block;
  background: var(--accent);
  color: #041014;
  font-weight: 700;
  font-size: 12px;
  padding: 4px 12px;
  border-radius: 20px;
  text-transform: uppercase;
  letter-spacing: 1px;
}
.hero h1 { font-size: 34px; margin: 16px 0 8px; line-height: 1.2; }
.hero p { color: var(--text-dim); font-size: 17px; }
.hero .meta { display: flex; gap: 24px; margin-top: 20px; flex-wrap: wrap; }
.hero .meta .item { font-size: 14px; }
.hero .meta .item b { color: var(--accent-2); display: block; font-size: 20px; }
.hero .meta .item span { color: var(--text-dim); }

/* ===== Sections ===== */
section { margin-bottom: 48px; scroll-margin-top: 24px; }
h2.section-title {
  font-size: 26px;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 2px solid var(--border);
  display: flex; align-items: center; gap: 10px;
}
h2.section-title .num {
  background: var(--accent); color: #041014;
  width: 32px; height: 32px; border-radius: 8px;
  display: inline-flex; align-items: center; justify-content: center;
  font-size: 16px; font-weight: 700; flex-shrink: 0;
}
h3 { font-size: 20px; margin: 28px 0 12px; color: var(--accent-2); }
h4 { font-size: 16px; margin: 20px 0 10px; color: var(--accent-3); }
p { margin-bottom: 14px; }

/* ===== Cards ===== */
.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 24px;
  margin-bottom: 20px;
}

.callout {
  border-left: 4px solid var(--accent);
  background: var(--bg-alt);
  padding: 16px 20px;
  border-radius: 0 8px 8px 0;
  margin: 18px 0;
}
.callout.warning { border-left-color: var(--accent-3); }
.callout.danger { border-left-color: var(--danger); }
.callout.tip { border-left-color: var(--accent-2); }
.callout .label {
  font-size: 12px; text-transform: uppercase; letter-spacing: 1px;
  font-weight: 700; margin-bottom: 6px;
}
.callout.warning .label { color: var(--accent-3); }
.callout.danger .label { color: var(--danger); }
.callout.tip .label { color: var(--accent-2); }
.callout .label.info { color: var(--accent); }

/* ===== Code ===== */
pre {
  background: var(--code-bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 18px 20px;
  overflow-x: auto;
  margin: 16px 0;
  font-size: 13.5px;
  line-height: 1.6;
}
code {
  font-family: "SF Mono", "JetBrains Mono", Menlo, Consolas, monospace;
}
p code, li code, td code {
  background: var(--bg-alt);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
  color: var(--purple);
  border: 1px solid var(--border);
}
pre code { color: var(--text); background: none; border: none; padding: 0; }

/* Simple syntax tokens */
.tok-kw { color: #ff7b72; }
.tok-fn { color: #d2a8ff; }
.tok-str { color: #a5d6ff; }
.tok-num { color: #79c0ff; }
.tok-com { color: #8b949e; font-style: italic; }
.tok-type { color: #7ee787; }

/* ===== Tables ===== */
table {
  width: 100%;
  border-collapse: collapse;
  margin: 18px 0;
  font-size: 14px;
  background: var(--bg-card);
  border-radius: 8px;
  overflow: hidden;
}
th, td { padding: 12px 16px; text-align: left; border-bottom: 1px solid var(--border); }
th { background: var(--bg-alt); color: var(--accent); font-weight: 600; }
tr:last-child td { border-bottom: none; }
td code { color: var(--purple); }

/* ===== Flashcards ===== */
.flashcard {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  margin-bottom: 12px;
  overflow: hidden;
}
.flashcard .q {
  padding: 16px 20px;
  cursor: pointer;
  font-weight: 600;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  transition: background .15s;
}
.flashcard .q:hover { background: var(--bg-alt); }
.flashcard .q .icon {
  color: var(--accent); font-size: 20px; transition: transform .2s; flex-shrink: 0;
}
.flashcard.open .q .icon { transform: rotate(45deg); }
.flashcard .q .qmark {
  color: var(--accent-3); font-weight: 700; margin-right: 8px;
}
.flashcard .a {
  max-height: 0;
  overflow: hidden;
  transition: max-height .3s ease, padding .3s ease;
  padding: 0 20px;
  border-top: 1px solid transparent;
}
.flashcard.open .a {
  max-height: 2000px;
  padding: 16px 20px;
  border-top: 1px solid var(--border);
}
.flashcard .a p:last-child { margin-bottom: 0; }

/* ===== Diagram box (ASCII fallback) ===== */
.diagram {
  background: var(--code-bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 20px;
  margin: 16px 0;
  font-family: "SF Mono", Menlo, monospace;
  font-size: 13px;
  white-space: pre;
  overflow-x: auto;
  color: var(--accent-2);
  line-height: 1.5;
}

/* ===== Pills / tags ===== */
.pill {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  margin: 2px;
}
.pill.stack { background: rgba(126,231,135,0.15); color: var(--accent-2); }
.pill.heap { background: rgba(255,123,114,0.15); color: var(--danger); }

/* ===== Nav buttons ===== */
.page-nav {
  display: flex; justify-content: space-between; gap: 16px;
  margin-top: 48px; padding-top: 24px; border-top: 1px solid var(--border);
}
.page-nav a {
  flex: 1;
  padding: 16px 20px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  text-decoration: none;
  color: var(--text);
  transition: all .15s;
}
.page-nav a:hover { border-color: var(--accent); background: var(--bg-alt); }
.page-nav a.next { text-align: right; }
.page-nav a .dir { font-size: 12px; color: var(--text-dim); display: block; }
.page-nav a .title { font-weight: 600; color: var(--accent); }
.page-nav a.disabled { opacity: 0.4; pointer-events: none; }

/* ===== Progress bar ===== */
.progress-bar {
  height: 6px; background: var(--bg-alt); border-radius: 3px; overflow: hidden; margin: 8px 0 4px;
}
.progress-bar .fill { height: 100%; background: linear-gradient(90deg, var(--accent), var(--accent-2)); }

/* ===== Key-value grid ===== */
.kv-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 12px; margin: 16px 0; }
.kv-grid .kv {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: 8px; padding: 16px;
}
.kv-grid .kv .k { font-size: 13px; color: var(--text-dim); }
.kv-grid .kv .v { font-size: 18px; font-weight: 700; color: var(--accent); margin-top: 4px; }

/* ===== week-grid (index landing) ===== */
.week-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 16px; margin: 24px 0; }
.week-card {
  background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius);
  padding: 20px 22px; text-decoration: none; color: var(--text); transition: all .15s; display: block;
}
.week-card:hover { border-color: var(--accent); background: var(--bg-alt); transform: translateY(-2px); }
.week-card .n { font-size: 12px; text-transform: uppercase; letter-spacing: 1px; color: var(--accent); font-weight: 700; }
.week-card h3 { margin: 8px 0 6px; color: var(--text); font-size: 18px; }
.week-card p { color: var(--text-dim); font-size: 14px; margin-bottom: 10px; }
.week-card .stars { font-size: 13px; }

/* ===== Responsive ===== */
@media (max-width: 900px) {
  .sidebar { display: none; }
  .content { padding: 24px 20px 80px; }
  .hero h1 { font-size: 26px; }
}

/* Toggle-all button */
.toggle-all {
  background: var(--bg-card); border: 1px solid var(--border); color: var(--accent);
  padding: 8px 16px; border-radius: 6px; cursor: pointer; font-size: 13px; margin-bottom: 16px;
  transition: all .15s;
}
.toggle-all:hover { background: var(--bg-alt); border-color: var(--accent); }

/* ===== Deep-dive components ===== */

/* Concept block wrapper */
.concept {
  border: 1px solid var(--border);
  border-radius: var(--radius);
  margin: 24px 0;
  overflow: hidden;
  background: var(--bg-card);
}
.concept > .concept-head {
  background: linear-gradient(135deg, rgba(0,173,216,0.12), transparent);
  padding: 16px 22px;
  border-bottom: 1px solid var(--border);
  font-size: 19px;
  font-weight: 700;
  color: var(--text);
  display: flex; align-items: center; gap: 10px;
}
.concept > .concept-body { padding: 6px 22px 18px; }

/* Labeled mini-blocks: why / how / purpose / pitfall / best */
.block {
  border-left: 4px solid var(--accent);
  background: var(--bg-alt);
  border-radius: 0 8px 8px 0;
  padding: 14px 18px;
  margin: 16px 0;
}
.block .block-label {
  font-size: 11px; text-transform: uppercase; letter-spacing: 1.2px;
  font-weight: 700; margin-bottom: 8px; display: flex; align-items: center; gap: 6px;
}
.block.why    { border-left-color: var(--purple); }
.block.why .block-label { color: var(--purple); }
.block.how    { border-left-color: var(--accent); }
.block.how .block-label { color: var(--accent); }
.block.purpose{ border-left-color: var(--accent-2); }
.block.purpose .block-label { color: var(--accent-2); }
.block.pitfall{ border-left-color: var(--danger); }
.block.pitfall .block-label { color: var(--danger); }
.block.best   { border-left-color: var(--accent-3); }
.block.best .block-label { color: var(--accent-3); }
.block p:last-child, .block ul:last-child { margin-bottom: 0; }
.block ul { margin-left: 18px; }
.block li { margin-bottom: 6px; }

/* Two-column comparison */
.grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin: 18px 0; }
.grid-2 .col {
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: 8px; padding: 16px 18px;
}
.grid-2 .col h4 { margin-top: 0; }
@media (max-width: 720px){ .grid-2 { grid-template-columns: 1fr; } }

/* Pros / Cons */
.proscons { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin: 18px 0; }
.proscons .pros, .proscons .cons {
  border-radius: 8px; padding: 16px 18px; border: 1px solid var(--border);
}
.proscons .pros { background: rgba(126,231,135,0.06); border-color: rgba(126,231,135,0.3); }
.proscons .cons { background: rgba(255,123,114,0.06); border-color: rgba(255,123,114,0.3); }
.proscons h4 { margin-top: 0; }
.proscons .pros h4 { color: var(--accent-2); }
.proscons .cons h4 { color: var(--danger); }
.proscons ul { margin-left: 18px; }
.proscons li { margin-bottom: 6px; font-size: 14.5px; }
@media (max-width: 720px){ .proscons { grid-template-columns: 1fr; } }

/* Numbered step flow */
.steps { counter-reset: step; margin: 18px 0; }
.steps .step {
  position: relative; padding: 4px 0 16px 44px; border-left: 2px solid var(--border);
  margin-left: 14px;
}
.steps .step:last-child { border-left-color: transparent; }
.steps .step::before {
  counter-increment: step; content: counter(step);
  position: absolute; left: -15px; top: 0;
  width: 28px; height: 28px; border-radius: 50%;
  background: var(--accent); color: #041014; font-weight: 700;
  display: flex; align-items: center; justify-content: center; font-size: 13px;
}
.steps .step .step-title { font-weight: 600; color: var(--accent-2); margin-bottom: 2px; }

/* TL;DR box */
.tldr {
  background: linear-gradient(135deg, rgba(210,168,255,0.1), transparent);
  border: 1px solid rgba(210,168,255,0.3);
  border-radius: var(--radius); padding: 18px 22px; margin: 18px 0;
}
.tldr .label { color: var(--purple); font-weight: 700; font-size: 12px; letter-spacing: 1px; text-transform: uppercase; margin-bottom: 8px; }
.tldr ul { margin-left: 18px; }
.tldr li { margin-bottom: 6px; }

/* Table caption */
.tbl-cap { font-size: 13px; color: var(--text-dim); margin: -8px 0 16px; font-style: italic; }
.subdivider { height: 1px; background: var(--border); margin: 28px 0; border: none; }

/* ============================================================
   DIAGRAM KIT — vẽ sơ đồ bằng HTML/CSS thay cho ASCII
   ============================================================ */
.dg {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 26px 24px;
  margin: 20px 0;
}
.dg-title { text-align: center; color: var(--text-dim); font-size: 13px; margin-bottom: 20px; font-style: italic; }
.dg-note { text-align: center; color: var(--text-dim); font-size: 12.5px; margin-top: 16px; line-height: 1.5; }

.dg-row { display: flex; align-items: center; justify-content: center; gap: 14px; flex-wrap: wrap; }
.dg-row.top { align-items: flex-start; }
.dg-col { display: flex; flex-direction: column; align-items: center; gap: 10px; }
.dg-col.left { align-items: flex-start; }
.dg-wrap { display: flex; flex-direction: column; align-items: center; gap: 12px; }
.dg-gap-lg { gap: 28px; }
.dg-grid { display: grid; gap: 16px; }

.node {
  display: inline-flex; flex-direction: column; align-items: center; justify-content: center;
  padding: 10px 16px; border-radius: 8px;
  border: 1px solid var(--border); background: var(--code-bg);
  font-size: 13px; font-weight: 600; text-align: center; min-width: 64px;
  line-height: 1.35;
}
.node .sub { font-size: 11px; color: var(--text-dim); font-weight: 400; margin-top: 3px; }
.node.accent { border-color: var(--accent); color: var(--accent); }
.node.green  { border-color: var(--accent-2); color: var(--accent-2); }
.node.orange { border-color: var(--accent-3); color: var(--accent-3); }
.node.red    { border-color: var(--danger); color: var(--danger); }
.node.purple { border-color: var(--purple); color: var(--purple); }
.node.muted  { color: var(--text-dim); border-style: dashed; }
.node.fill-accent { background: var(--accent); color: #041014; border-color: var(--accent); }
.node.fill-green  { background: rgba(126,231,135,0.15); border-color: var(--accent-2); color: var(--accent-2); }
.node.wide { min-width: 150px; }
.node.sm { padding: 6px 10px; font-size: 12px; min-width: 40px; }
.node.big { padding: 16px 22px; font-size: 15px; }

.arrow { display: flex; align-items: center; justify-content: center; color: var(--text-dim); position: relative; }
.arrow-h { flex: 1 1 40px; min-width: 40px; height: 2px; background: var(--border); position: relative; }
.arrow-h::after { content: ""; position: absolute; right: -1px; top: -4px; border: 5px solid transparent; border-left-color: var(--border); }
.arrow-h.rev::after { right: auto; left: -1px; border-left-color: transparent; border-right-color: var(--border); }
.arrow-v { width: 2px; height: 30px; background: var(--border); position: relative; margin: 2px 0; }
.arrow-v::after { content: ""; position: absolute; bottom: -1px; left: -4px; border: 5px solid transparent; border-top-color: var(--border); }
.arrow-v.rev::after { bottom: auto; top: -1px; border-top-color: transparent; border-bottom-color: var(--border); }
.arrow-v.tall { height: 48px; }
.arrow-h.acc, .arrow-v.acc { background: var(--accent); }
.arrow-h.acc::after { border-left-color: var(--accent); }
.arrow-v.acc::after { border-top-color: var(--accent); }
.arrow-h.grn, .arrow-v.grn { background: var(--accent-2); }
.arrow-h.grn::after { border-left-color: var(--accent-2); }
.arrow-v.grn::after { border-top-color: var(--accent-2); }
.arrow-h.red, .arrow-v.red { background: var(--danger); }
.arrow-h.red::after { border-left-color: var(--danger); }
.arrow-v.red::after { border-top-color: var(--danger); }
.arrow-lbl { font-size: 11px; color: var(--text-dim); padding: 0 6px; white-space: nowrap; }

.cells { display: inline-flex; gap: 4px; padding: 4px; border-radius: 8px; }
.cell {
  width: 36px; height: 36px; border: 1px solid var(--border); border-radius: 6px;
  background: var(--code-bg); display: flex; align-items: center; justify-content: center;
  font-family: "SF Mono", Menlo, monospace; font-size: 12px; color: var(--text);
}
.cell.fill { border-color: var(--accent-2); color: var(--accent-2); }
.cell.empty { color: var(--text-dim); border-style: dashed; }
.cell.hot { border-color: var(--accent-3); color: var(--accent-3); }

.dg-group {
  border: 1px dashed var(--border); border-radius: 10px; padding: 16px;
  display: flex; flex-direction: column; align-items: center; gap: 10px;
}
.dg-group .grp-label { font-size: 11px; text-transform: uppercase; letter-spacing: 1px; color: var(--text-dim); }

.dg-legend { display: flex; gap: 18px; justify-content: center; flex-wrap: wrap; margin-top: 16px; }
.dg-legend .lg { display: inline-flex; align-items: center; gap: 6px; font-size: 12px; color: var(--text-dim); }
.dg-legend .sw { width: 14px; height: 14px; border-radius: 3px; border: 1px solid var(--border); }
.dg-legend .sw.accent { border-color: var(--accent); background: rgba(0,173,216,0.2); }
.dg-legend .sw.green  { border-color: var(--accent-2); background: rgba(126,231,135,0.2); }
.dg-legend .sw.orange { border-color: var(--accent-3); background: rgba(255,166,87,0.2); }
.dg-legend .sw.red    { border-color: var(--danger); background: rgba(255,123,114,0.2); }

@media (max-width: 720px) {
  .dg-row.scroll { overflow-x: auto; justify-content: flex-start; }
}
```

---

## APP.JS  (ghi vào `html/assets/app.js`)

```js
// ===== Flashcard toggle =====
document.addEventListener('click', function (e) {
  const q = e.target.closest('.flashcard .q');
  if (q) {
    q.parentElement.classList.toggle('open');
  }
});

// ===== Toggle all flashcards =====
function toggleAll(btn) {
  const cards = document.querySelectorAll('.flashcard');
  const anyClosed = Array.from(cards).some(c => !c.classList.contains('open'));
  cards.forEach(c => c.classList.toggle('open', anyClosed));
  btn.textContent = anyClosed ? 'Collapse all' : 'Expand all';
}

// ===== Active nav highlight on scroll =====
const sections = document.querySelectorAll('section[id]');
const navLinks = document.querySelectorAll('.sidebar nav a[href^="#"]');

function onScroll() {
  let current = '';
  sections.forEach(sec => {
    const top = sec.getBoundingClientRect().top;
    if (top <= 120) current = sec.id;
  });
  navLinks.forEach(link => {
    link.classList.toggle('active', link.getAttribute('href') === '#' + current);
  });
}
window.addEventListener('scroll', onScroll);
window.addEventListener('load', onScroll);
```

---

## INDEX TEMPLATE  (ghi vào `html/index.html`)

```html
<!DOCTYPE html>
<html lang="vi">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{TOPIC}} Study Hub</title>
<link rel="stylesheet" href="assets/style.css">
</head>
<body>
<div class="layout">
  <aside class="sidebar">
    <div class="brand">
      <a href="index.html"><div class="logo">{{TOPIC}} knowledge</div></a>
      <div class="sub">Study Hub</div>
    </div>
    <h2>All Weeks</h2>
    <nav>
      <a href="week1.html">Week 1 · {{...}}</a>
      <!-- ...tất cả tuần... -->
    </nav>
  </aside>
  <main class="content">
    <div class="hero">
      <span class="week-badge">{{TOPIC}}</span>
      <h1>{{Tiêu đề bộ tài liệu}}</h1>
      <p>{{Mô tả ngắn}}</p>
      <div class="meta">
        <div class="item"><b>{{n}}</b><span>Tuần</span></div>
        <div class="item"><b>{{x}}+</b><span>Code examples</span></div>
        <div class="item"><b>{{y}}+</b><span>Flashcards</span></div>
      </div>
    </div>
    <div class="week-grid">
      <a class="week-card" href="week1.html">
        <div class="n">Week 1</div><h3>{{Tên}}</h3>
        <p>{{Mô tả 1 dòng}}</p>
        <div class="stars">⭐⭐⭐⭐⭐</div>
      </a>
      <!-- ...mỗi tuần một card... -->
    </div>
  </main>
</div>
<script src="assets/app.js"></script>
</body>
</html>
```

---

## PAGE TEMPLATE  (ghi vào `html/weekN.html`)

```html
<!DOCTYPE html>
<html lang="vi">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Week N – {{TÊN}} (Deep Dive) | {{TOPIC}}</title>
<link rel="stylesheet" href="assets/style.css">
</head>
<body>
<div class="layout">
  <aside class="sidebar">
    <div class="brand">
      <a href="index.html"><div class="logo">{{TOPIC}} knowledge</div></a>
      <div class="sub">Deep Dive</div>
    </div>
    <h2>Week N · Contents</h2>
    <nav>
      <a href="#overview">Overview &amp; TL;DR</a>
      <a href="#s1">1. {{Mục 1}}</a>
      <a href="#flashcards">Flashcards</a>
    </nav>
    <h2>All Weeks</h2>
    <nav>
      <!-- liệt kê MỌI tuần; trang hiện tại thêm class="active" -->
      <a href="week1.html">Week 1 · {{...}}</a>
    </nav>
  </aside>

  <main class="content">
    <div class="hero">
      <span class="week-badge">Week N · Deep Dive</span>
      <h1>{{Tiêu đề tuần}}</h1>
      <p>{{Mô tả}}</p>
      <div class="meta">
        <div class="item"><b>★★★★</b><span>Priority</span></div>
        <div class="item"><b>{{n}}</b><span>Chủ đề</span></div>
        <div class="item"><b>~{{h}}h</b><span>Đề xuất</span></div>
      </div>
    </div>

    <section id="overview">
      <h2 class="section-title"><span class="num">i</span> Overview &amp; TL;DR</h2>
      <div class="tldr"><div class="label">TL;DR</div><ul><li>...</li></ul></div>
    </section>

    <section id="s1">
      <h2 class="section-title"><span class="num">1</span> {{Mục 1}}</h2>
      <div class="block why"><div class="block-label">Vì sao</div><p>...</p></div>
      <pre><code><span class="tok-kw">func</span> <span class="tok-fn">demo</span>() { <span class="tok-com">// ...</span> }</code></pre>
      <table><tr><th>Cột A</th><th>Cột B</th></tr><tr><td>...</td><td>...</td></tr></table>
      <div class="block pitfall"><div class="block-label">Bẫy</div><ul><li>...</li></ul></div>
    </section>

    <section id="flashcards">
      <h2 class="section-title"><span class="num">?</span> Flashcards</h2>
      <button class="toggle-all" onclick="toggleAll(this)">Expand all</button>
      <div class="flashcard">
        <div class="q"><span><span class="qmark">Q</span>{{câu hỏi}}?</span><span class="icon">+</span></div>
        <div class="a"><p>{{trả lời}}</p></div>
      </div>
    </section>

    <div class="page-nav">
      <a href="{{prev}}"><span class="dir">← Trước</span><span class="title">{{prev title}}</span></a>
      <a href="{{next}}" class="next"><span class="dir">Tiếp theo →</span><span class="title">{{next title}}</span></a>
    </div>
  </main>
</div>
<script src="assets/app.js"></script>
</body>
</html>
```

---

## DIAGRAM KIT — ví dụ dùng

Vẽ luồng ngang (A → B → C):

```html
<div class="dg">
  <div class="dg-title">Request flow</div>
  <div class="dg-row scroll">
    <div class="node accent">Client</div>
    <div class="arrow-h acc"></div>
    <div class="node">API</div>
    <div class="arrow-h acc"></div>
    <div class="node green">DB</div>
  </div>
  <div class="dg-legend">
    <span class="lg"><span class="sw accent"></span> entry</span>
    <span class="lg"><span class="sw green"></span> storage</span>
  </div>
</div>
```

Buffer/queue bằng `.cells`/`.cell`; nhóm bằng `.dg-group`; luồng dọc bằng `.arrow-v`.

---

## Quy tắc bắt buộc (checklist HTML)

- [ ] Tạo đủ `assets/style.css` + `assets/app.js` (copy nguyên văn ở trên) TRƯỚC khi tạo trang.
- [ ] Mỗi `<section>` có `id` khớp với link `href="#id"` trong sidebar "Contents".
- [ ] Sidebar "All Weeks" có mặt & đồng bộ trên mọi trang; trang hiện tại gắn `class="active"`.
- [ ] Escape `&amp;` `&lt;` `&gt;` trong nội dung; code tô màu bằng `.tok-*`.
- [ ] Ưu tiên component (`.block`, `.tldr`, `.grid-2`, diagram kit) thay cho HTML thô/ASCII.
- [ ] Số `<section>` == số `</section>`; link prev/next trong `.page-nav` đúng.
- [ ] `index.html` có `.week-grid` với một `.week-card` cho mỗi trang.
