---
inclusion: manual
---

# Study Material Generator

Hướng dẫn để Kiro sinh **bộ tài liệu ôn tập / phỏng vấn** cho một ngôn ngữ hoặc framework
bất kỳ, theo đúng cấu trúc và phong cách của repo Go hiện có (`study/`).

## Mục đích

Tái tạo trải nghiệm học tập gồm 3 lớp cho một chủ đề mới (vd: Rust, Python, Node.js, React,
Kubernetes, PostgreSQL, System Design...):

1. **Master plan** — roadmap chia tuần + question bank + final checklist.
2. **Study materials theo tuần** — `notes.md` (deep dive), `flashcards.md` (Q&A), `code_examples/` (chạy được).
3. **HTML study hub** — trang web deep-dive dùng chung design system tối màu, có sidebar, flashcard toggle, sơ đồ vẽ bằng CSS.

## Khi nào dùng

Khi người dùng yêu cầu "tạo tài liệu học/ôn X", "làm study plan cho X", hoặc "generate tài liệu
kiểu như repo Go này cho X". Kích hoạt bằng cách tham chiếu file steering này (`#study-material-generator`).

---

## Quy trình sinh tài liệu (workflow)

Luôn dùng `todo_list` để theo dõi. Trình tự chuẩn:

1. **Làm rõ mục tiêu** (nếu chưa rõ): ngôn ngữ/framework nào, cấp độ (junior/mid/**senior**),
   mục đích (phỏng vấn / lên tay / ôn nhanh), số tuần mong muốn (mặc định 6–8).
2. **Nghiên cứu roadmap**: dùng web search (vd roadmap.sh, docs chính thức) để lấy danh sách topic
   chuẩn. Diễn giải lại nội dung (tuân thủ licensing, không copy > 30 từ liên tiếp), có trích nguồn.
3. **Chốt roadmap chia tuần**: mỗi tuần một cụm chủ đề, gán priority ⭐. Ưu tiên phần "core/hiểu sâu"
   trước, ecosystem/tooling sau.
4. **Tạo master plan** (`<topic>_plan.md`) trước để thống nhất khung.
5. **Sinh từng tuần**: với mỗi tuần tạo `notes.md`, `flashcards.md`, và `code_examples/`.
6. **Tạo HTML hub**: copy assets, tạo `index.html`, rồi `weekN.html` cho từng tuần.
7. **Tạo checklist + README**.
8. **Verify**: format & chạy thử code examples; kiểm tra HTML (thẻ cân bằng, anchor khớp nav).

Có thể sinh dần từng tuần thay vì tất cả cùng lúc nếu chủ đề lớn — hỏi người dùng muốn làm hết
hay theo từng phần.

---

## Cấu trúc thư mục chuẩn

```
<topic>-study/
├── README.md                          # tổng quan + cây thư mục + cách dùng
├── <topic>_plan.md                    # master plan (roadmap, question bank, final checklist)
├── checklist/
│   └── weekly_checklist.md            # tracking tiến độ theo tuần
├── week1-<slug>/                      # vd: week1-rust-ownership
│   ├── notes.md
│   ├── flashcards.md
│   └── code_examples/                 # file chạy được, mỗi file standalone
├── week2-<slug>/ ...
└── html/
    ├── index.html                     # study hub landing (grid các week card)
    ├── week1.html ...                 # trang deep-dive từng tuần
    └── assets/
        ├── style.css                  # COPY từ repo Go (rebrand qua --accent)
        └── app.js                     # COPY từ repo Go (flashcard toggle + nav scroll)
```

Đặt tên folder tuần: `weekN-<kebab-slug>`. Code examples đặt tên mô tả (`ownership_move.rs`,
`async_await.py`...). Nếu ngôn ngữ có module system (Go modules, Cargo, npm) thì mỗi file
code_example là **snippet standalone chạy độc lập** (giống repo Go: nhiều `main` cùng thư mục,
chạy từng file), trừ khi người dùng muốn một project build được hoàn chỉnh.

---

## Quy ước Markdown

### Master plan (`<topic>_plan.md`)

- `# 1. Mục tiêu` — đối tượng, cấp độ, kết quả kỳ vọng.
- `# 2. Roadmap tổng quan` — bảng `| Tuần | Chủ đề | Priority |` với ⭐⭐⭐⭐⭐.
- `# 3..N Week X – <tên>` — mỗi tuần: danh sách topic dạng checkbox `- [ ]`, nhóm theo `##`.
- `# Question Bank` — câu hỏi phỏng vấn nhóm theo chủ đề (`## <chủ đề>` + `- [ ] câu hỏi?`).
- `# Final Preparation Checklist` — checklist tổng.
- (Tùy chọn) `# Recommended Interview Mindset`, mapping từ ngôn ngữ người dùng đã biết.

### `notes.md` (deep dive từng tuần)

- Mở đầu bằng callout blockquote `>` nêu phạm vi tuần.
- Heading đánh số: `## 1. <Chủ đề>`, `## 2. ...`.
- Dùng **bảng** để so sánh, **code block** có comment giải thích, **danh sách** cho quy tắc.
- Mỗi chủ đề nên có: *khái niệm → cơ chế bên dưới (why/how) → ví dụ code → bẫy thường gặp →
  góc nhìn senior*.
- Ngăn cách các mục lớn bằng `---`.

### `flashcards.md`

Định dạng cố định (app không parse file này, nhưng giữ nhất quán để dễ chuyển sang HTML):

```
### Q: <câu hỏi>?
**A:** <câu trả lời ngắn gọn, chính xác, có "vì sao">

---
```

Nhóm theo chủ đề bằng `## <chủ đề>`.

### `weekly_checklist.md`

`## Week N – <tên>` → các nhóm `### <chủ đề>` → checkbox `- [ ]`.
Kết thúc mỗi tuần bằng `### Code Examples Completed` liệt kê file.

### `README.md`

Tổng quan mục tiêu + **cây thư mục** (ASCII, có comment mỗi mục) + cách dùng + cách chạy code.

---

## Quy ước Code examples

- Ưu tiên **chạy được thật** bằng stdlib; hạn chế dependency ngoài. Nếu buộc phải có dep
  (gRPC, driver DB...), viết snippet minh hoạ + ghi rõ cần cài gì.
- Comment bằng tiếng Việt giải thích *tại sao*, không chỉ *cái gì*.
- Đầu file ghi cách chạy (vd `// Chạy: go run file.go`, `# Chạy: python file.py`).
- **Verify trước khi giao**: format bằng formatter chuẩn của ngôn ngữ (gofmt / rustfmt / black /
  prettier) và chạy thử. Với file test, chạy test runner để chắc chắn pass. Dọn file tạm.

---

## HTML Study Hub — Design System

### Assets dùng chung

Copy nguyên `html/assets/style.css` và `html/assets/app.js` từ repo Go. Đây là dark theme
hoàn chỉnh; **rebrand** cho ngôn ngữ mới chỉ cần đổi biến `--accent` (và tuỳ chọn `--accent-2/3`)
trong `:root` của `style.css` — ví dụ Rust dùng cam `#dea584`, Python xanh `#4b8bbe`.

`app.js` cung cấp: click câu hỏi để mở/đóng flashcard (`.flashcard.open`), nút `toggleAll(this)`,
highlight nav theo scroll (dựa vào `section[id]` + `.sidebar nav a[href^="#"]`). Không cần sửa.

### Bộ class có sẵn (dùng lại, đừng tự chế class mới)

| Nhóm | Class |
|---|---|
| Layout | `.layout`, `.sidebar`, `.brand`/`.logo`/`.sub`, `.content` |
| Hero | `.hero`, `.week-badge`, `.meta`/`.item` |
| Section | `.section-title` + `.num`, `h3`, `h4` |
| Callout mini-block | `.block` + biến thể `.why` `.how` `.purpose` `.pitfall` `.best`, kèm `.block-label` |
| Box nhấn mạnh | `.tldr` + `.label`; `.callout` + `.warning/.danger/.tip` |
| So sánh | `.grid-2 > .col`; `.proscons > .pros/.cons` |
| Bảng | `<table>` (style sẵn) + `.tbl-cap` |
| Code token | `.tok-kw` `.tok-fn` `.tok-str` `.tok-num` `.tok-com` `.tok-type` |
| Flashcard | `.flashcard > .q(.qmark,.icon) + .a`; nút `.toggle-all` |
| Sơ đồ (diagram kit) | `.dg`/`.dg-title`/`.dg-note`, `.dg-row`/`.dg-col`, `.node`(+màu), `.arrow-h`/`.arrow-v`(+`.acc/.grn/.red`, `.rev`), `.cells`/`.cell`, `.dg-group`, `.dg-legend` |
| Steps | `.steps > .step > .step-title` |
| Điều hướng | `.page-nav > a(.next/.disabled) > .dir + .title` |

Ưu tiên **vẽ sơ đồ bằng diagram kit** (node + arrow) thay cho ASCII art để đẹp và responsive.

### Skeleton trang tuần (`weekN.html`)

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
      <!-- ...mỗi section một link, href = #id của section... -->
      <a href="#flashcards">Flashcards</a>
    </nav>
    <h2>All Weeks</h2>
    <nav>
      <!-- liệt kê TẤT CẢ tuần; gắn class="active" cho trang hiện tại -->
      <a href="week1.html">Week 1 · {{...}}</a>
      <!-- ... -->
    </nav>
  </aside>

  <main class="content">
    <div class="hero">
      <span class="week-badge">Week N · Deep Dive</span>
      <h1>{{Tiêu đề tuần}}</h1>
      <p>{{Mô tả 1–2 câu}}</p>
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
      <pre><code>...code có token span...</code></pre>
      <div class="block pitfall"><div class="block-label">Bẫy</div><ul><li>...</li></ul></div>
    </section>

    <!-- ...các section khác... -->

    <section id="flashcards">
      <h2 class="section-title"><span class="num">?</span> Flashcards</h2>
      <button class="toggle-all" onclick="toggleAll(this)">Expand all</button>
      <div class="flashcard">
        <div class="q"><span><span class="qmark">Q</span>{{câu hỏi}}?</span><span class="icon">+</span></div>
        <div class="a"><p>{{trả lời}}</p></div>
      </div>
      <!-- ...thêm flashcard... -->
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

### `index.html` (landing)

Dùng `.hero` với `.meta` (số tuần / code examples / flashcards) và một `.week-grid` chứa các
`.week-card` (mỗi card: `.n` số tuần, `h3` tên, `p` mô tả, `.stars` ⭐). Mỗi card link tới `weekN.html`.

### Quy tắc HTML

- `id` của `<section>` phải khớp với `href="#id"` trong sidebar "Contents".
- Sidebar "All Weeks" phải liệt kê **mọi tuần** và có mặt trên **mọi trang** (khi thêm tuần mới,
  cập nhật nav ở tất cả trang).
- Escape ký tự trong code: `&amp;` `&lt;` `&gt;`. Tô màu code bằng các `.tok-*` span.
- Sau khi tạo/sửa: kiểm tra số `<section>` = `</section>`, anchor khớp nav.

---

## Nguyên tắc chất lượng nội dung

- **Độ sâu senior**: giải thích cơ chế bên dưới (memory model, runtime, internals), trade-off,
  và "khi nào dùng cái gì" — không dừng ở cú pháp.
- **Chính xác kỹ thuật** hơn dài dòng. Nếu không chắc, tra cứu docs chính thức, đừng bịa.
- **Song ngữ**: giải thích tiếng Việt, giữ thuật ngữ/tên API tiếng Anh.
- **Bám thực tế phỏng vấn**: mỗi chủ đề gắn với câu hay bị hỏi + bẫy thường gặp.
- **Tuân thủ nguồn**: khi lấy từ web, diễn giải lại (không copy > 30 từ liên tiếp), trích link nguồn.

## Checklist hoàn thành (tự kiểm trước khi báo xong)

- [ ] Master plan có roadmap + question bank + final checklist.
- [ ] Mỗi tuần đủ `notes.md` + `flashcards.md` + `code_examples/`.
- [ ] Code examples đã format và **chạy/test thử OK**.
- [ ] HTML: `index.html` + đủ `weekN.html`, assets đã copy, `--accent` đã rebrand.
- [ ] Nav "All Weeks" đồng bộ trên mọi trang; anchor khớp; thẻ `<section>` cân bằng.
- [ ] `checklist/weekly_checklist.md` và `README.md` phản ánh đúng nội dung thực tế.
