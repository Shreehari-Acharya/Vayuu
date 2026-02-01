# 📄 Skill: URL To Markdown Scraper

**Location:** `skills/url_to_markdown.md`

**Standard:** Use this SOP to convert any public webpage URL into clean, distraction-free Markdown content suitable for LLM processing.

**Important:** Always store generated files in summaries/ to maintain organization.

**Names:** use better relevant names for output files based on the webpage title or topic.

---


### 🛡️ Stage 1: HTML Fetch

The agent must use a masked **Headless Chrome** instance. This prevents the website (especially those on Vercel or Cloudflare) from identifying the agent as a bot.

**Command:**

```bash
google-chrome --headless --disable-gpu --dump-dom \
  --user-agent="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36" \
  --disable-blink-features=AutomationControlled \
  "[TARGET_URL]" > summaries/raw.html

```

---

### 🏗️ Stage 2: Structural Conversion

The agent uses **Pandoc** to convert the raw HTML into a "Strict" Markdown format. This automatically discards 90% of web clutter like navigation bars, sidebars, and CSS classes.

**Command:**

```bash
pandoc summaries/raw.html -f html-native_divs-native_spans -t markdown_strict-raw_html --wrap=none -o summaries/clean_content.md

```

* **`-t markdown_strict-raw_html`**: Forces the removal of all modern HTML attributes and tags.
* **`--wrap=none`**: Keeps text in a single continuous flow for better LLM context.

---

### ✂️ Stage 3: Surgical Noise Removal

The final stage uses **sed** with the **in-place (`-i`)** flag to delete remaining visual distractions that Pandoc might have missed, specifically image syntax and leftover metadata.

**Commands:**

```bash
# Delete all Markdown image tags: ![alt](url)
sed -i -E 's/!\[[^]]*\]\([^)]*\)//g' summaries/clean_content.md

```

---

### 📋 SOP Summary Table

| Action | Tool | Purpose |
| --- | --- | --- |
| **Fetch** | Chrome Stealth | Bypasses **429 Errors** and **WAFs**. |
| **Convert** | Pandoc Strict | Strips **Tailwind/CSS** and structural junk. |
| **Scrub** | sed -i | Deletes **Image URLs** to save tokens. |

