#!/usr/bin/env python3
"""Generate benchmark fixtures for Thumbnail Forge."""
import os
import struct
import zlib
import gzip
import tarfile
import zipfile
import json
import random
import string
import io
import subprocess
import sys

BASE = "tests/bench"
os.makedirs(BASE, exist_ok=True)

def ensure_dir(path):
    os.makedirs(path, exist_ok=True)

def write_file(path, content):
    ensure_dir(os.path.dirname(path))
    if isinstance(content, str):
        content = content.encode()
    with open(path, 'wb') as f:
        f.write(content)
    size = os.path.getsize(path)
    print(f"  {path} ({size:,} bytes)")
    return path

# ============================================================
# IMAGES
# ============================================================
def create_png(width, height, path):
    """Create a PNG using ffmpeg."""
    subprocess.run([
        'ffmpeg', '-f', 'lavfi', '-i', f'testsrc=size={width}x{height}:rate=1',
        '-frames:v', '1', '-y', path
    ], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    print(f"  {path} ({os.path.getsize(path):,} bytes)")
    return path

def create_jpeg(width, height, path):
    """Create a JPEG using ffmpeg."""
    # Generate a single frame with ffmpeg, extract as JPEG
    subprocess.run([
        'ffmpeg', '-f', 'lavfi', '-i', f'color=c=0x4488cc:s={width}x{height}:d=0.04',
        '-frames:v', '1', '-q:v', '2', '-y', path
    ], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    print(f"  {path} ({os.path.getsize(path):,} bytes)")
    return path

def create_gif(width, height, path):
    """Create a GIF using ffmpeg."""
    subprocess.run([
        'ffmpeg', '-f', 'lavfi', '-i', f'color=c=0x884422:s={width}x{height}:d=0.04',
        '-frames:v', '1', '-f', 'gif', '-y', path
    ], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    print(f"  {path} ({os.path.getsize(path):,} bytes)")
    return path

def create_webp(width, height, path):
    """Create WebP using ffmpeg."""
    subprocess.run([
        'ffmpeg', '-f', 'lavfi', '-i', f'color=c=0x22cc88:s={width}x{height}:d=0.04',
        '-frames:v', '1', '-y', path
    ], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    print(f"  {path} ({os.path.getsize(path):,} bytes)")
    return path

def create_bmp(width, height, path):
    """Create a BMP file manually."""
    row_size = (width * 3 + 3) & ~3
    pixel_data_size = row_size * height
    file_size = 54 + pixel_data_size
    
    header = struct.pack('<2sIHHI', b'BM', file_size, 0, 0, 54)
    info_header = struct.pack('<IIIHHIIIIII',
        40, width, height, 1, 24, 0, pixel_data_size, 2835, 2835, 0, 0)
    
    with open(path, 'wb') as f:
        f.write(header + info_header)
        for y in range(height):
            for x in range(width):
                r = (x * 255) // max(width, 1)
                g = (y * 255) // max(height, 1)
                b = 128
                f.write(bytes([b, g, r]))
            # Padding
            padding = row_size - (width * 3)
            f.write(b'\x00' * padding)
    
    print(f"  {path} ({os.path.getsize(path):,} bytes)")
    return path

def create_tiff(width, height, path):
    """Create a TIFF by converting from PNG using Go's image package."""
    # First create a PNG, then convert to TIFF using a small Go program
    png_path = path.replace('.tiff', '.png')
    subprocess.run([
        'ffmpeg', '-f', 'lavfi', '-i', f'testsrc=size={width}x{height}:rate=1',
        '-frames:v', '1', '-y', png_path
    ], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    
    # Convert using Go
    go_code = f'''package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"golang.org/x/image/tiff"
)

func main() {{
	f, _ := os.Open("{png_path}")
	defer f.Close()
	img, _ := png.Decode(f)
	
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(bounds)
	for y := 0; y < bounds.Dy(); y++ {{
		for x := 0; x < bounds.Dx(); x++ {{
			nrgba.Set(x, y, color.NRGBAModel.Convert(img.At(x, y)))
		}}
	}}
	
	out, _ := os.Create("{path}")
	defer out.Close()
	tiff.Encode(out, nrgba, nil)
}}
'''
    go_file = path + ".go"
    with open(go_file, 'w') as f:
        f.write(go_code)
    subprocess.run(['go', 'run', go_file], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, cwd='.')
    os.remove(go_file)
    if os.path.exists(png_path):
        os.remove(png_path)
    
    if os.path.exists(path):
        print(f"  {path} ({os.path.getsize(path):,} bytes)")
    else:
        print(f"  {path} FAILED")
    return path

def generate_images():
    print("\n=== IMAGES ===")
    ensure_dir(f"{BASE}/images")
    
    # Small: 64x64, Medium: 512x512, Large: 1024x1024
    create_png(64, 64, f"{BASE}/images/small.png")
    create_png(512, 512, f"{BASE}/images/medium.png")
    create_png(1024, 1024, f"{BASE}/images/large.png")
    
    create_jpeg(64, 64, f"{BASE}/images/small.jpg")
    create_jpeg(512, 512, f"{BASE}/images/medium.jpg")
    create_jpeg(1024, 1024, f"{BASE}/images/large.jpg")
    
    create_gif(64, 64, f"{BASE}/images/small.gif")
    create_gif(256, 256, f"{BASE}/images/medium.gif")
    
    create_webp(64, 64, f"{BASE}/images/small.webp")
    create_webp(256, 256, f"{BASE}/images/medium.webp")
    
    create_bmp(64, 64, f"{BASE}/images/small.bmp")
    create_bmp(256, 256, f"{BASE}/images/medium.bmp")
    
    create_tiff(64, 64, f"{BASE}/images/small.tiff")
    create_tiff(256, 256, f"{BASE}/images/medium.tiff")

# ============================================================
# VIDEO
# ============================================================
def generate_video(width, height, duration, path, codec='libx264', fmt='mp4'):
    """Generate a video using ffmpeg."""
    if fmt == 'webm':
        cmd = [
            'ffmpeg', '-f', 'lavfi',
            '-i', f'testsrc=size={width}x{height}:rate=25:duration={duration}',
            '-c:v', 'libvpx',
            '-y', path
        ]
    else:
        cmd = [
            'ffmpeg', '-f', 'lavfi',
            '-i', f'testsrc=size={width}x{height}:rate=25:duration={duration}',
            '-c:v', codec,
            '-pix_fmt', 'yuv420p',
            '-y', path
        ]
    subprocess.run(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    if not os.path.exists(path):
        print(f"  {path} FAILED")
        return None
    print(f"  {path} ({os.path.getsize(path):,} bytes)")
    return path

def generate_videos():
    print("\n=== VIDEO ===")
    ensure_dir(f"{BASE}/video")
    
    # Small: 160x120 2s, Medium: 640x480 5s, Large: 1280x720 10s
    generate_video(160, 120, 2, f"{BASE}/video/small.mp4")
    generate_video(640, 480, 5, f"{BASE}/video/medium.mp4")
    generate_video(1280, 720, 10, f"{BASE}/video/large.mp4")
    
    generate_video(160, 120, 2, f"{BASE}/video/small.mov")
    generate_video(160, 120, 2, f"{BASE}/video/small.mkv")
    generate_video(160, 120, 2, f"{BASE}/video/small.webm", fmt='webm')

# ============================================================
# AUDIO
# ============================================================
def generate_audio(duration, path, codec='libmp3lame', fmt='mp3'):
    """Generate audio using ffmpeg."""
    cmd = [
        'ffmpeg', '-f', 'lavfi',
        '-i', f'sine=frequency=440:duration={duration}',
        '-c:a', codec,
        '-y', path
    ]
    subprocess.run(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    if not os.path.exists(path):
        print(f"  {path} FAILED")
        return None
    print(f"  {path} ({os.path.getsize(path):,} bytes)")
    return path

def generate_audios():
    print("\n=== AUDIO ===")
    ensure_dir(f"{BASE}/audio")
    
    # Small: 1s, Medium: 10s, Large: 60s
    generate_audio(1, f"{BASE}/audio/small.mp3")
    generate_audio(10, f"{BASE}/audio/medium.mp3")
    generate_audio(1, f"{BASE}/audio/small.wav", codec='pcm_s16le')
    generate_audio(1, f"{BASE}/audio/small.flac", codec='flac')
    generate_audio(1, f"{BASE}/audio/small.ogg", codec='libvorbis')

# ============================================================
# PDF
# ============================================================
def generate_pdf(pages, path):
    """Generate a PDF using LibreOffice from a text file."""
    # Create a text file with content
    txt_path = path.replace('.pdf', '.txt')
    content = ""
    for p in range(pages):
        content += f"Page {p+1}\n\n"
        for i in range(50):
            content += f"This is line {i} on page {p+1}. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n"
        content += "\f"  # form feed = page break
    
    write_file(txt_path, content)
    
    # Convert to PDF using LibreOffice
    subprocess.run([
        'libreoffice', '--headless', '--convert-to', 'pdf',
        '--outdir', os.path.dirname(path), txt_path
    ], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=30)
    
    # LibreOffice names it .pdf
    generated = txt_path.replace('.txt', '.pdf')
    if os.path.exists(generated) and generated != path:
        os.rename(generated, path)
    
    if os.path.exists(txt_path):
        os.remove(txt_path)
    
    if not os.path.exists(path):
        print(f"  {path} FAILED")
        return None
    print(f"  {path} ({os.path.getsize(path):,} bytes)")
    return path

def generate_pdfs():
    print("\n=== PDF ===")
    ensure_dir(f"{BASE}/pdf")
    
    generate_pdf(1, f"{BASE}/pdf/small.pdf")
    generate_pdf(10, f"{BASE}/pdf/medium.pdf")
    generate_pdf(50, f"{BASE}/pdf/large.pdf")

# ============================================================
# OFFICE
# ============================================================
def generate_docx(paragraphs, path):
    """Generate a DOCX file."""
    content_types = '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>'''
    
    rels = '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>'''
    
    body_content = ""
    for i in range(paragraphs):
        body_content += f'''    <w:p>
      <w:r>
        <w:t>Paragraph {i+1}: Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</w:t>
      </w:r>
    </w:p>
'''
    
    document = f'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
{body_content}  </w:body>
</w:document>'''
    
    with zipfile.ZipFile(path, 'w', zipfile.ZIP_DEFLATED) as z:
        z.writestr('[Content_Types].xml', content_types)
        z.writestr('_rels/.rels', rels)
        z.writestr('word/document.xml', document)
    
    print(f"  {path} ({os.path.getsize(path):,} bytes)")
    return path

def generate_xlsx(rows, path):
    """Generate an XLSX file."""
    content_types = '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
</Types>'''
    
    rels = '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>'''
    
    workbook = '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheets>
    <sheet name="Sheet1" sheetId="1" r:id="rId1" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"/>
  </sheets>
</workbook>'''
    
    sheet_data = '<sheetData>'
    for r in range(1, rows + 1):
        sheet_data += f'<row r="{r}"><c r="A{r}" t="inlineStr"><is><t>Name{r}</t></is></c><c r="B{r}"><v>{r * 10}</v></c><c r="C{r}"><v>{r * 20}</v></c></row>'
    sheet_data += '</sheetData>'
    
    sheet = f'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  {sheet_data}
</worksheet>'''
    
    with zipfile.ZipFile(path, 'w', zipfile.ZIP_DEFLATED) as z:
        z.writestr('[Content_Types].xml', content_types)
        z.writestr('_rels/.rels', rels)
        z.writestr('xl/workbook.xml', workbook)
        z.writestr('xl/worksheets/sheet1.xml', sheet)
    
    print(f"  {path} ({os.path.getsize(path):,} bytes)")
    return path

def generate_pptx(slides, path):
    """Generate a PPTX file."""
    content_types = '''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>'''
    
    rels_body = ''
    pres_body = ''
    
    for s in range(1, slides + 1):
        content_types += f'\n  <Override PartName="/ppt/slides/slide{s}.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>'
        rels_body += f'\n  <Relationship Id="rId{s}" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide{s}.xml"/>'
        pres_body += f'<sldId id="{s}" r:id="rId{s}"/>'
    
    content_types += '\n</Types>'
    
    rels = f'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId0" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>{rels_body}
</Relationships>'''
    
    presentation = f'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sldIdLst>{pres_body}</sldIdLst>
</p:presentation>'''
    
    with zipfile.ZipFile(path, 'w', zipfile.ZIP_DEFLATED) as z:
        z.writestr('[Content_Types].xml', content_types)
        z.writestr('_rels/.rels', rels)
        z.writestr('ppt/presentation.xml', presentation)
        for s in range(1, slides + 1):
            slide = f'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
  <p:cSld><p:spTree>
    <p:sp>
      <p:nvSpPr><p:cNvPr id="1" name="Title{s}"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr>
      <p:spPr><a:xfrm><a:off x="100" y="100"/><a:ext cx="5000" cy="500"/></a:xfrm><a:prstGeom prst="rect"/></p:spPr>
      <p:txBody><a:bodyPr/><a:lstStyle/><a:p><a:r><a:t>Slide {s} - Thumbnail Forge Benchmark</a:t></a:r></a:p></p:txBody>
    </p:sp>
  </p:spTree></p:cSld>
</p:sld>'''
            z.writestr(f'ppt/slides/slide{s}.xml', slide)
    
    print(f"  {path} ({os.path.getsize(path):,} bytes)")
    return path

def generate_office():
    print("\n=== OFFICE ===")
    ensure_dir(f"{BASE}/office")
    
    generate_docx(5, f"{BASE}/office/small.docx")
    generate_docx(100, f"{BASE}/office/medium.docx")
    generate_xlsx(10, f"{BASE}/office/small.xlsx")
    generate_pptx(3, f"{BASE}/office/small.pptx")

# ============================================================
# CODE
# ============================================================
def generate_go_code(lines, path):
    code = '''package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"
	"math"
	"sort"
	"time"
)

// DataItem represents a data record
type DataItem struct {
	ID       int
	Name     string
	Value    float64
	Category string
	Tags     []string
	Created  time.Time
}

// Processor handles batch processing of data items
type Processor struct {
	items    []DataItem
	results  map[string][]DataItem
	stats    map[string]float64
}

func NewProcessor() *Processor {
	return &Processor{
		items:   make([]DataItem, 0),
		results: make(map[string][]DataItem),
		stats:   make(map[string]float64),
	}
}

func (p *Processor) AddItem(item DataItem) {
	p.items = append(p.items, item)
}

func (p *Processor) Process() {
	for _, item := range p.items {
		category := item.Category
		p.results[category] = append(p.results[category], item)
		p.stats[category] += item.Value
	}
}

func (p *Processor) GetStats() map[string]float64 {
	return p.stats
}

func (p *Processor) SortByValue() {
	sort.Slice(p.items, func(i, j int) bool {
		return p.items[i].Value > p.items[j].Value
	})
}

func main() {
	proc := NewProcessor()
'''
    for i in range(lines - 50):
        code += f'\tproc.AddItem(DataItem{{ID: {i}, Name: "item_{i}", Value: {i * 1.5:.2f}, Category: "cat_{i % 5}", Tags: []string{{"tag1", "tag2"}}, Created: time.Now()}})\n'
    
    code += '''\tproc.Process()
\tproc.SortByValue()
\tstats := proc.GetStats()
\tfor k, v := range stats {
\t\tfmt.Printf("%s: %.2f\\n", k, v)
\t}
\t_ = os.Args
\t_ = strings.Builder{}
\t_ = strconv.Itoa
	_ = math.Pi
}
'''
    write_file(path, code)

def generate_py_code(lines, path):
    code = '''#!/usr/bin/env python3
"""Benchmark sample Python file."""

import os
import sys
import json
import time
import math
import random
import hashlib
import datetime
from collections import defaultdict
from typing import List, Dict, Optional, Tuple


class DataProcessor:
    """Processes data items in batches."""
    
    def __init__(self, batch_size: int = 100):
        self.batch_size = batch_size
        self.items: List[dict] = []
        self.results: Dict[str, List[dict]] = defaultdict(list)
    
    def add_item(self, item: dict) -> None:
        self.items.append(item)
    
    def process_batch(self) -> Dict[str, List[dict]]:
        for item in self.items:
            category = item.get("category", "unknown")
            self.results[category].append(item)
        return dict(self.results)
    
    def get_stats(self) -> Dict[str, float]:
        stats = {}
        for cat, items in self.results.items():
            total = sum(i.get("value", 0) for i in items)
            stats[cat] = total / len(items) if items else 0
        return stats


def main():
    proc = DataProcessor(batch_size=50)
'''
    for i in range(lines - 40):
        code += f'    proc.add_item({{"id": {i}, "name": "item_{i}", "value": {i * 1.5:.2f}, "category": "cat_{i % 5}"}})\n'
    
    code += '''    results = proc.process_batch()
    stats = proc.get_stats()
    for k, v in stats.items():
        print(f"{k}: {v:.2f}")


if __name__ == "__main__":
    main()
'''
    write_file(path, code)

def generate_js_code(lines, path):
    code = '''/**
 * Benchmark sample JavaScript file.
 */

const fs = require('fs');
const path = require('path');
const crypto = require('crypto');

class DataProcessor {
    constructor(batchSize = 100) {
        this.batchSize = batchSize;
        this.items = [];
        this.results = new Map();
    }
    
    addItem(item) {
        this.items.push(item);
    }
    
    processBatch() {
        for (const item of this.items) {
            const category = item.category || 'unknown';
            if (!this.results.has(category)) {
                this.results.set(category, []);
            }
            this.results.get(category).push(item);
        }
        return this.results;
    }
    
    getStats() {
        const stats = {};
        for (const [cat, items] of this.results) {
            const total = items.reduce((sum, i) => sum + (i.value || 0), 0);
            stats[cat] = items.length > 0 ? total / items.length : 0;
        }
        return stats;
    }
}

async function main() {
    const proc = new DataProcessor(50);
'''
    for i in range(lines - 40):
        code += f'    proc.addItem({{id: {i}, name: "item_{i}", value: {i * 1.5:.2f}, category: "cat_{i % 5}"}});\n'
    
    code += '''    proc.processBatch();
    const stats = proc.getStats();
    for (const [k, v] of Object.entries(stats)) {
        console.log(`${k}: ${v.toFixed(2)}`);
    }
}

main().catch(console.error);
'''
    write_file(path, code)

def generate_ts_code(lines, path):
    code = '''/**
 * Benchmark sample TypeScript file.
 */

interface DataItem {
    id: number;
    name: string;
    value: number;
    category: string;
    tags?: string[];
}

interface ProcessResult {
    [category: string]: DataItem[];
}

interface Stats {
    [category: string]: number;
}

class DataProcessor {
    private batchSize: number;
    private items: DataItem[] = [];
    private results: ProcessResult = {};

    constructor(batchSize: number = 100) {
        this.batchSize = batchSize;
    }

    addItem(item: DataItem): void {
        this.items.push(item);
    }

    processBatch(): ProcessResult {
        for (const item of this.items) {
            const category = item.category || 'unknown';
            if (!this.results[category]) {
                this.results[category] = [];
            }
            this.results[category].push(item);
        }
        return this.results;
    }

    getStats(): Stats {
        const stats: Stats = {};
        for (const [cat, items] of Object.entries(this.results)) {
            const total = items.reduce((sum, i) => sum + i.value, 0);
            stats[cat] = items.length > 0 ? total / items.length : 0;
        }
        return stats;
    }
}

function main(): void {
    const proc = new DataProcessor(50);
'''
    for i in range(lines - 45):
        code += f'    proc.addItem({{id: {i}, name: "item_{i}", value: {i * 1.5:.2f}, category: "cat_{i % 5}"}});\n'
    
    code += '''    proc.processBatch();
    const stats = proc.getStats();
    for (const [k, v] of Object.entries(stats)) {
        console.log(`${k}: ${v.toFixed(2)}`);
    }
}

main();
'''
    write_file(path, code)

def generate_rust_code(lines, path):
    code = '''// Benchmark sample Rust file.

use std::collections::HashMap;
use std::time::Instant;

struct DataItem {
    id: u32,
    name: String,
    value: f64,
    category: String,
}

struct DataProcessor {
    items: Vec<DataItem>,
    results: HashMap<String, Vec<DataItem>>,
}

impl DataProcessor {
    fn new() -> Self {
        DataProcessor {
            items: Vec::new(),
            results: HashMap::new(),
        }
    }

    fn add_item(&mut self, item: DataItem) {
        self.items.push(item);
    }

    fn process_batch(&mut self) {
        for item in self.items.drain(..) {
            let category = item.category.clone();
            self.results.entry(category).or_insert_with(Vec::new).push(item);
        }
    }

    fn get_stats(&self) -> HashMap<String, f64> {
        let mut stats = HashMap::new();
        for (cat, items) in &self.results {
            let total: f64 = items.iter().map(|i| i.value).sum();
            let avg = if items.is_empty() { 0.0 } else { total / items.len() as f64 };
            stats.insert(cat.clone(), avg);
        }
        stats
    }
}

fn main() {
    let mut proc = DataProcessor::new();
'''
    for i in range(lines - 50):
        code += f'    proc.add_item(DataItem {{ id: {i}, name: "item_{i}".to_string(), value: {i * 1.5:.2f}, category: "cat_{i % 5}".to_string() }});\n'
    
    code += '''    proc.process_batch();
    let stats = proc.get_stats();
    for (k, v) in &stats {
        println!("{}: {:.2}", k, v);
    }
}
'''
    write_file(path, code)

def generate_java_code(lines, path):
    code = '''import java.util.*;
import java.time.*;

public class DataProcessor {
    private int batchSize;
    private List<DataItem> items;
    private Map<String, List<DataItem>> results;
    
    public DataProcessor(int batchSize) {
        this.batchSize = batchSize;
        this.items = new ArrayList<>();
        this.results = new HashMap<>();
    }
    
    public void addItem(DataItem item) {
        items.add(item);
    }
    
    public void processBatch() {
        for (DataItem item : items) {
            String category = item.getCategory();
            results.computeIfAbsent(category, k -> new ArrayList<>()).add(item);
        }
    }
    
    public Map<String, Double> getStats() {
        Map<String, Double> stats = new HashMap<>();
        for (Map.Entry<String, List<DataItem>> entry : results.entrySet()) {
            double total = 0;
            for (DataItem i : entry.getValue()) total += i.getValue();
            stats.put(entry.getKey(), total / entry.getValue().size());
        }
        return stats;
    }
    
    public static void main(String[] args) {
        DataProcessor proc = new DataProcessor(50);
'''
    for i in range(lines - 40):
        code += f'        proc.addItem(new DataItem({i}, "item_{i}", {i * 1.5:.2f}, "cat_{i % 5}"));\n'
    
    code += '''        proc.processBatch();
        Map<String, Double> stats = proc.getStats();
        for (Map.Entry<String, Double> e : stats.entrySet()) {
            System.out.printf("%s: %.2f%n", e.getKey(), e.getValue());
        }
    }
}

class DataItem {
    private int id;
    private String name;
    private double value;
    private String category;
    
    public DataItem(int id, String name, double value, String category) {
        this.id = id;
        this.name = name;
        this.value = value;
        this.category = category;
    }
    
    public String getCategory() { return category; }
    public double getValue() { return value; }
}
'''
    write_file(path, code)

def generate_c_code(lines, path):
    code = '''#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <time.h>

typedef struct {
    int id;
    char name[64];
    double value;
    int category;
} DataItem;

typedef struct {
    DataItem *items;
    int count;
    int capacity;
} DataProcessor;

void processor_init(DataProcessor *p, int capacity) {
    p->items = malloc(capacity * sizeof(DataItem));
    p->count = 0;
    p->capacity = capacity;
}

void processor_add(DataProcessor *p, DataItem item) {
    if (p->count < p->capacity) {
        p->items[p->count++] = item;
    }
}

void processor_process(DataProcessor *p) {
    double sums[10] = {0};
    int counts[10] = {0};
    for (int i = 0; i < p->count; i++) {
        int cat = p->items[i].category;
        sums[cat] += p->items[i].value;
        counts[cat]++;
    }
    for (int i = 0; i < 10; i++) {
        if (counts[i] > 0) {
            printf("cat_%d: %.2f\\n", i, sums[i] / counts[i]);
        }
    }
}

int main() {
    DataProcessor proc;
    processor_init(&proc, 10000);
'''
    for i in range(lines - 50):
        code += f'    processor_add(&proc, (DataItem){{{i}, "item_{i}", {i * 1.5:.2f}, {i % 5}}});\n'
    
    code += '''    processor_process(&proc);
    free(proc.items);
    return 0;
}
'''
    write_file(path, code)

def generate_code():
    print("\n=== CODE ===")
    ensure_dir(f"{BASE}/code")
    
    # Small: ~50 lines, Medium: ~500 lines, Large: ~2000 lines
    generate_go_code(50, f"{BASE}/code/small.go")
    generate_go_code(500, f"{BASE}/code/medium.go")
    generate_go_code(2000, f"{BASE}/code/large.go")
    
    generate_py_code(50, f"{BASE}/code/small.py")
    generate_js_code(50, f"{BASE}/code/small.js")
    generate_ts_code(50, f"{BASE}/code/small.ts")
    generate_rust_code(50, f"{BASE}/code/small.rs")
    generate_java_code(50, f"{BASE}/code/small.java")
    generate_c_code(50, f"{BASE}/code/small.c")

# ============================================================
# TEXT
# ============================================================
def generate_text():
    print("\n=== TEXT ===")
    ensure_dir(f"{BASE}/text")
    
    # TXT
    small = "Hello, World!\nThis is a test file.\n" * 5
    medium = "Lorem ipsum dolor sit amet, consectetur adipiscing elit.\n" * 200
    large = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n" * 2000
    write_file(f"{BASE}/text/small.txt", small)
    write_file(f"{BASE}/text/medium.txt", medium)
    write_file(f"{BASE}/text/large.txt", large)
    
    # JSON
    small_json = json.dumps([{"id": i, "name": f"item_{i}", "value": i * 1.5} for i in range(10)], indent=2)
    medium_json = json.dumps([{"id": i, "name": f"item_{i}", "value": i * 1.5, "tags": ["a", "b", "c"], "nested": {"x": i, "y": i * 2}} for i in range(200)], indent=2)
    write_file(f"{BASE}/text/small.json", small_json)
    write_file(f"{BASE}/text/medium.json", medium_json)
    
    # XML
    xml_small = '<?xml version="1.0"?>\n<root>\n' + '\n'.join(f'  <item id="{i}"><name>item_{i}</name><value>{i * 1.5}</value></item>' for i in range(10)) + '\n</root>'
    write_file(f"{BASE}/text/small.xml", xml_small)
    
    # YAML
    yaml_small = 'items:\n' + '\n'.join(f'  - id: {i}\n    name: item_{i}\n    value: {i * 1.5}\n    tags:\n      - tag1\n      - tag2' for i in range(10))
    write_file(f"{BASE}/text/small.yaml", yaml_small)
    
    # CSV
    csv_small = "id,name,value,category\n" + '\n'.join(f'{i},item_{i},{i * 1.5:.2f},cat_{i % 5}' for i in range(20))
    csv_medium = "id,name,value,category,tags,created\n" + '\n'.join(f'{i},item_{i},{i * 1.5:.2f},cat_{i % 5},"tag1;tag2",2026-01-{(i % 28) + 1:02d}' for i in range(500))
    write_file(f"{BASE}/text/small.csv", csv_small)
    write_file(f"{BASE}/text/medium.csv", csv_medium)

# ============================================================
# MARKDOWN
# ============================================================
def generate_markdown():
    print("\n=== MARKDOWN ===")
    ensure_dir(f"{BASE}/markdown")
    
    small = "# Test Document\n\nThis is a **test** markdown file.\n\n## Section 1\n\n- Item 1\n- Item 2\n- Item 3\n\n## Section 2\n\nSome `code` here.\n"
    medium = "# Benchmark Markdown\n\n" + '\n'.join(f"## Section {i}\n\nLorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n\n- Item {i}.1\n- Item {i}.2\n- Item {i}.3\n\n```python\ndef hello():\n    print('hello world')\n```\n" for i in range(20))
    large = "# Large Markdown Document\n\n" + '\n'.join(f"## Section {i}\n\nLorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.\n\n### Subsection {i}.1\n\n- First item\n- Second item\n- Third item with **bold** text\n- Fourth item with *italic* text\n\n### Subsection {i}.2\n\n```go\nfunc process(data []byte) error {{\n    fmt.Println(string(data))\n    return nil\n}}\n```\n" for i in range(100))
    write_file(f"{BASE}/markdown/small.md", small)
    write_file(f"{BASE}/markdown/medium.md", medium)
    write_file(f"{BASE}/markdown/large.md", large)

# ============================================================
# ARCHIVES
# ============================================================
def generate_archives():
    print("\n=== ARCHIVES ===")
    ensure_dir(f"{BASE}/archives")
    
    # Create some files to archive
    temp_dir = f"{BASE}/archives/_temp"
    ensure_dir(temp_dir)
    
    # Small set of files
    for i in range(5):
        write_file(f"{temp_dir}/file_{i}.txt", f"Content of file {i}\n" * 10)
    
    # Small ZIP
    small_zip = f"{BASE}/archives/small.zip"
    with zipfile.ZipFile(small_zip, 'w', zipfile.ZIP_DEFLATED) as z:
        for i in range(5):
            z.write(f"{temp_dir}/file_{i}.txt", f"file_{i}.txt")
    print(f"  {small_zip} ({os.path.getsize(small_zip):,} bytes)")
    
    # Medium ZIP (more files)
    for i in range(5, 50):
        write_file(f"{temp_dir}/file_{i}.txt", f"Content of file {i}\n" * 20)
    medium_zip = f"{BASE}/archives/medium.zip"
    with zipfile.ZipFile(medium_zip, 'w', zipfile.ZIP_DEFLATED) as z:
        for i in range(50):
            z.write(f"{temp_dir}/file_{i}.txt", f"file_{i}.txt")
    print(f"  {medium_zip} ({os.path.getsize(medium_zip):,} bytes)")
    
    # Large ZIP (big files)
    for i in range(50, 100):
        write_file(f"{temp_dir}/file_{i}.txt", f"Content of file {i}\n" * 100)
    large_zip = f"{BASE}/archives/large.zip"
    with zipfile.ZipFile(large_zip, 'w', zipfile.ZIP_DEFLATED) as z:
        for i in range(100):
            z.write(f"{temp_dir}/file_{i}.txt", f"file_{i}.txt")
    print(f"  {large_zip} ({os.path.getsize(large_zip):,} bytes)")
    
    # TAR
    small_tar = f"{BASE}/archives/small.tar"
    with tarfile.open(small_tar, 'w') as t:
        for i in range(5):
            t.add(f"{temp_dir}/file_{i}.txt", f"file_{i}.txt")
    print(f"  {small_tar} ({os.path.getsize(small_tar):,} bytes)")
    
    # TAR.GZ
    small_targz = f"{BASE}/archives/small.tar.gz"
    with tarfile.open(small_targz, 'w:gz') as t:
        for i in range(5):
            t.add(f"{temp_dir}/file_{i}.txt", f"file_{i}.txt")
    print(f"  {small_targz} ({os.path.getsize(small_targz):,} bytes)")
    
    # Cleanup temp
    import shutil
    shutil.rmtree(temp_dir)

# ============================================================
# MAIN
# ============================================================
if __name__ == '__main__':
    print("Generating benchmark fixtures for Thumbnail Forge...")
    print(f"Base directory: {BASE}")
    
    generate_images()
    generate_videos()
    generate_audios()
    generate_pdfs()
    generate_office()
    generate_code()
    generate_text()
    generate_markdown()
    generate_archives()
    
    # Count all files
    total = 0
    total_size = 0
    for root, dirs, files in os.walk(BASE):
        for f in files:
            total += 1
            total_size += os.path.getsize(os.path.join(root, f))
    
    print(f"\n{'='*60}")
    print(f"Total files: {total}")
    print(f"Total size: {total_size:,} bytes ({total_size / 1024 / 1024:.1f} MB)")
    print(f"{'='*60}")
