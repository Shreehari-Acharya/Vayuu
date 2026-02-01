# 🎨 Skill: Visual Document Engineering

**Location:** `skills/visual_doc_eng.md`

**Standard:** Use this guide to determine the toolchain for creating professional-grade assets.

**Important:** Always store generated files in docs/ to maintain organization.

---

## 📑 1. PDF Generation: Technical vs. Aesthetic

You have two distinct paths depending on the document's purpose.

### Path A: The Technical Path (Pandoc)

**Use for:** Documentation, technical manuals, whitepapers, and academic reports.

* **Why:** Pandoc handles complex tables, cross-references, and Table of Contents (TOC) with structural integrity.
* **Command:** `pandoc input.md -o docs/output.pdf --toc --highlight-style=tango`
* **Agent Tip:** Ensure Markdown is clean. Use `#` for levels. Pandoc will handle the "pro" layout automatically.

### Path B: The Aesthetic Path (HTML ➔ Chrome)

**Use for:** Resumes, portfolios, invoices, and "Personal Use" files that must be **really beautiful**.

* **Why:** Chrome Headless renders modern CSS (Gradients, Glassmorphism, Custom Fonts) that other engines fail to see.
* **The Workflow:** 1. Generate a "Standout" HTML file with an internal `<style>` block.
2. Convert using Chrome Headless.
* **Command:** `google-chrome --headless --disable-gpu --print-to-pdf="docs/out.pdf" --no-pdf-header-footer input.html`

---

## 📄 2. Word Documents (DOCX)

**Standard:** Descent, professional, and editable. No reference template needed.

* **Why:** Pandoc has excellent "Sensible Defaults." It creates a clean, standard Word doc that is perfectly readable and easy for a human to finish editing.
* **Strategy:** Don't worry about complex styling here; focus on **semantic hierarchy**. Use proper H1, H2, and Bullet points so the DOCX "Styles" pane works correctly for the user.
* **Command:** `pandoc input.md -o docs/output.docx`

---

## 📽️ 3. Presentations (PPTX)

**Standard:** "Executive-Level" & "Professional." **Marp is the only choice.**

* **Why:** Marp is the industry leader for AI-to-Slide conversion. It allows you to use Markdown to control professional layouts that standard PPTX generators cannot achieve.

### 💡 The "Beautiful" Slide Rules:

1. **Directives:** Always start with `marp: true` and a theme like but not limited to `theme: uncover`.
2. **Visual Balance:** Use the "Background Image" trick for 50/50 split slides:
`![bg right:45%](image_url)`
3. **Typography:** Use the `header` and `footer` directives to keep a consistent brand presence on every slide.

### 🛠 Execution:

`marp --pptx input.md -o docs/output.pptx --allow-local-files`

---

## 💎 Critical Execution Checklist 

1. **Format Detection:** Check the user's intent. If they say "make it look great," default to **Path B (Chrome)** for PDFs or **Marp** for PPTs.
2. **HTML Styling:** When using Chrome, always include `@page { margin: 0; }` or `@media print` CSS to ensure the beauty isn't ruined by default browser margins.
3. **File Cleanup:** After rendering the final `.pdf` or `.pptx`, you may delete the temporary `.html` or `.md` source files unless the user asks for them.

---
