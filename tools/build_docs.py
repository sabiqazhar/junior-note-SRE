#!/usr/bin/env python3
"""Generate readable .html sibling for every .md in the repo.

Usage:
    /home/pluto/Documents/.venv/bin/python tools/build_docs.py

- Walks: ./production-reliability-lab + ./README.md (root)
- For each foo.md writes foo.html next to it (same dir)
- Template: dark theme matching index.html, responsive, TOC, copy-button.
- Rewrites internal href="*.md" -> href="*.html" (+ anchors preserved).
- Safe to re-run (idempotent).
"""
import re
import html as htmlmod
from pathlib import Path

import markdown

ROOT = Path(__file__).resolve().parent.parent
SCAN_ROOTS = [ROOT / "production-reliability-lab", ROOT / "README.md"]

CSS = """
:root{--bg:#0d1117;--bg2:#161b22;--bg3:#1c2128;--line:#30363d;--fg:#e6edf3;--fg2:#9198a1;--accent:#58a6ff;--accent2:#3fb950;--warn:#d29922;--radius:10px}
*{box-sizing:border-box}html,body{margin:0;padding:0}
body{background:var(--bg);color:var(--fg);font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;font-size:15.5px;line-height:1.7}
a{color:var(--accent);text-decoration:none}a:hover{text-decoration:underline}
.top{position:sticky;top:0;z-index:10;background:#0d1117ee;backdrop-filter:blur(6px);border-bottom:1px solid var(--line);padding:10px 20px;display:flex;gap:12px;align-items:center;font-size:13.5px}
.top .crumb{color:var(--fg2);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.btn{background:var(--bg3);border:1px solid var(--line);color:var(--fg);padding:6px 13px;border-radius:8px;font-size:13px;white-space:nowrap}
.btn:hover{border-color:var(--accent);color:var(--accent);text-decoration:none}
.wrap{max-width:860px;margin:0 auto;padding:28px 22px 90px}
h1{font-size:26px;line-height:1.3;border-bottom:1px solid var(--line);padding-bottom:12px}
h2{font-size:21px;margin-top:38px;border-bottom:1px solid var(--line);padding-bottom:8px}
h3{font-size:17px;margin-top:28px}h4{font-size:15px;margin-top:22px}
p,li{font-size:15px}li{margin-bottom:6px}
code{background:var(--bg3);padding:1px 6px;border-radius:5px;font-size:.87em;font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace}
pre{background:var(--bg2);border:1px solid var(--line);border-radius:var(--radius);padding:14px 16px;overflow-x:auto;font-size:13.5px;line-height:1.55;position:relative}
pre code{background:none;padding:0}
.copy{position:absolute;top:8px;right:8px;background:var(--bg3);border:1px solid var(--line);color:var(--fg2);border-radius:6px;font-size:11px;padding:3px 9px;cursor:pointer}
.copy:hover{color:var(--accent);border-color:var(--accent)}
table{border-collapse:collapse;width:100%;font-size:13.8px;margin:16px 0;display:block;overflow-x:auto}
th,td{border:1px solid var(--line);padding:9px 12px;text-align:left}th{background:var(--bg3)}
blockquote{background:var(--bg2);border-left:3px solid var(--accent);margin:16px 0;padding:12px 18px;border-radius:0 var(--radius) var(--radius) 0;color:var(--fg)}
blockquote p{margin:8px 0}
hr{border:none;border-top:1px solid var(--line);margin:28px 0}
img{max-width:100%;border-radius:8px;border:1px solid var(--line)}
.toc{background:var(--bg2);border:1px solid var(--line);border-radius:var(--radius);padding:14px 18px;margin:0 0 26px;font-size:13.8px}
.toc summary{cursor:pointer;font-weight:700}
.toc ul{margin:8px 0;padding-left:20px}.toc li{margin-bottom:4px}
.task-list-item{list-style:none}.task-list-item input{margin-right:8px;accent-color:var(--accent2)}
.foot{margin-top:50px;padding-top:16px;border-top:1px solid var(--line);color:var(--fg2);font-size:12.5px;display:flex;justify-content:space-between;gap:10px;flex-wrap:wrap}
.codehilite .hll{background:#272822}.codehilite .c{color:#75715e}.codehilite .k{color:#66d9ef}.codehilite .s{color:#e6db74}.codehilite .n,.codehilite .nb,.codehilite .nf{color:#a6e22e}.codehilite .m{color:#ae81ff}
@media(max-width:640px){.wrap{padding:20px 15px 70px}h1{font-size:22px}}
"""

JS = """
document.querySelectorAll('pre').forEach(function(pre){
  var b=document.createElement('button');b.className='copy';b.textContent='copy';
  b.onclick=function(){var t=pre.innerText.replace(/\\ncopy$/,'');navigator.clipboard.writeText(t).then(function(){b.textContent='copied!';setTimeout(function(){b.textContent='copy'},1200)})};
  pre.appendChild(b);
});
"""

TEMPLATE = """<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{title}</title>
<style>{css}</style>
</head>
<body>
<div class="top"><a class="btn" href="{back_href}">&#8592; Roadmap</a><span class="crumb">{crumb}</span></div>
<div class="wrap">
{toc}
{body}
<div class="foot"><span>{crumb}</span><span><a href="{back_href}">Kembali ke roadmap</a></span></div>
</div>
<script>{js}</script>
</body>
</html>
"""

MD_LINK_RE = re.compile(r'href="([^"]*?)\.md((?:#[^"]*)?)"')


def rewrite_md_links(html: str) -> str:
    return MD_LINK_RE.sub(r'href="\1.html\2"', html)


def find_md_files():
    files = []
    for r in SCAN_ROOTS:
        if r.is_file() and r.suffix == ".md":
            files.append(r)
        elif r.is_dir():
            files.extend(sorted(r.rglob("*.md")))
    # skip node_modules/.git
    files = [f for f in files if ".git/" not in str(f)]
    return files


def build_one(md_path: Path) -> Path:
    text = md_path.read_text(encoding="utf-8")
    md = markdown.Markdown(
        extensions=["extra", "toc", "sane_lists"],
        extension_configs={"toc": {"permalink": True, "baselevel": 1}},
    )
    body = md.convert(text)
    body = rewrite_md_links(body)
    toc = f'<details class="toc" open><summary>Daftar isi</summary>{md.toc}</details>' if getattr(md, "toc", "").strip() and len(md.toc) > 50 else ""
    # title: first h1 or filename
    m = re.search(r"^#\s+(.+)$", text, re.M)
    title = m.group(1).strip() if m else md_path.stem.replace("-", " ").replace("_", " ")
    rel = md_path.relative_to(ROOT)
    # back link: relative from md dir to ROOT/index.html
    back_href = "../" * (len(rel.parts) - 1) + "index.html"
    if md_path.parent == ROOT:
        back_href = "index.html"
    out = md_path.with_suffix(".html")
    out.write_text(
        TEMPLATE.format(
            title=htmlmod.escape(title),
            css=CSS,
            js=JS,
            toc=toc,
            body=body,
            crumb=htmlmod.escape(str(rel)),
            back_href=back_href,
        ),
        encoding="utf-8",
    )
    return out


def main():
    files = find_md_files()
    outs = [build_one(f) for f in files]
    print(f"Rendered {len(outs)} markdown files -> html")
    for o in outs[:5]:
        print("  ", o.relative_to(ROOT))
    if len(outs) > 5:
        print(f"  ... +{len(outs)-5} more")


if __name__ == "__main__":
    main()
