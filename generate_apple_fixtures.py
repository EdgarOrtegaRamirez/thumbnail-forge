#!/usr/bin/env python3
"""Generate Apple format test fixtures for Thumbnail Forge."""
import os
import struct
import subprocess
import zipfile
import io

FIXTURES_DIR = os.path.join(os.path.dirname(__file__), "tests", "fixtures")
os.makedirs(FIXTURES_DIR, exist_ok=True)

def run(cmd, **kwargs):
    kwargs.setdefault('stdout', subprocess.DEVNULL)
    kwargs.setdefault('stderr', subprocess.DEVNULL)
    subprocess.run(cmd, **kwargs)

def create_heic(path):
    """Create a HEIC file using heif-enc."""
    png_path = path + ".src.png"
    run(['ffmpeg', '-f', 'lavfi', '-i', 'testsrc=size=128x128:rate=1',
         '-frames:v', '1', '-y', png_path])
    run(['heif-enc', png_path, '-o', path])
    os.unlink(png_path)
    print(f"  {path} ({os.path.getsize(path)} bytes)")

def create_avif(path):
    """Create an AVIF file using heif-enc with AV1 codec."""
    png_path = path + ".src.png"
    run(['ffmpeg', '-f', 'lavfi', '-i', 'testsrc=size=128x128:rate=1',
         '-frames:v', '1', '-y', png_path])
    run(['heif-enc', '-A', png_path, '-o', path])  # -A for AVIF
    os.unlink(png_path)
    print(f"  {path} ({os.path.getsize(path)} bytes)")

def create_icns(path):
    """Create a minimal ICNS file with an embedded PNG."""
    png_path = path + ".src.png"
    run(['ffmpeg', '-f', 'lavfi', '-i', 'testsrc=size=128x128:rate=1',
         '-frames:v', '1', '-y', png_path])
    with open(png_path, 'rb') as f:
        png_data = f.read()
    os.unlink(png_path)

    # ICNS header: 'icns' + total file size (big-endian uint32)
    total_size = 8 + 8 + len(png_data)
    header = b'icns' + struct.pack('>I', total_size)
    # Icon entry: 'ic07' (128x128 PNG) + entry size + PNG data
    icon_entry = b'ic07' + struct.pack('>I', 8 + len(png_data)) + png_data

    with open(path, 'wb') as f:
        f.write(header + icon_entry)
    print(f"  {path} ({os.path.getsize(path)} bytes)")

def create_alac(path):
    """Create an ALAC audio file (M4A container with ALAC codec)."""
    run(['ffmpeg', '-f', 'lavfi', '-i', 'sine=frequency=440:duration=2',
         '-c:a', 'alac', '-y', path])
    print(f"  {path} ({os.path.getsize(path)} bytes)")

def create_aiff(path):
    """Create an AIFF audio file."""
    run(['ffmpeg', '-f', 'lavfi', '-i', 'sine=frequency=440:duration=2',
         '-c:a', 'pcm_s16be', '-y', path])
    print(f"  {path} ({os.path.getsize(path)} bytes)")

def create_prores(path):
    """Create a ProRes video in a MOV container."""
    run(['ffmpeg', '-f', 'lavfi', '-i', 'testsrc=size=160x120:rate=25:duration=2',
         '-c:v', 'prores_ks', '-y', path])
    print(f"  {path} ({os.path.getsize(path)} bytes)")

def create_dmg(path):
    """Create a minimal DMG file (just the 'koly' trailer)."""
    # A real DMG has a UDIF trailer with 'koly' magic at the end.
    # We create a minimal valid structure for detection testing.
    koly_header = b'koly'  # magic
    koly_header += struct.pack('>I', 512)  # trailer size
    koly_header += b'\x00' * 504  # padding to 512 bytes
    with open(path, 'wb') as f:
        f.write(koly_header)
    print(f"  {path} ({os.path.getsize(path)} bytes)")

def create_ipa(path):
    """Create a minimal IPA file (ZIP with iOS app structure)."""
    with zipfile.ZipFile(path, 'w', zipfile.ZIP_DEFLATED) as zf:
        # IPA structure: Payload/App.app/...
        zf.writestr('Payload/TestApp.app/Info.plist',
                    '<?xml version="1.0"?>\n<plist version="1.0">\n'
                    '<dict><key>CFBundleName</key><string>TestApp</string></dict>\n</plist>')
        zf.writestr('Payload/TestApp.app/TestApp', b'\x00' * 1024)
        zf.writestr('Payload/TestApp.app/icon.png',
                    b'\x89PNG\r\n\x1a\n' + b'\x00' * 100)  # fake icon
    print(f"  {path} ({os.path.getsize(path)} bytes)")

def create_pages(path):
    """Create a minimal Apple Pages file (ZIP-based)."""
    with zipfile.ZipFile(path, 'w', zipfile.ZIP_DEFLATED) as zf:
        zf.writestr('index.xml', '<?xml version="1.0"?>\n<plist version="1.0">\n'
                    '<dict><key>DocumentType</key><string>Publishing</string></dict>\n</plist>')
        zf.writestr('preview.jpg', b'\xFF\xD8\xFF' + b'\x00' * 50)
    print(f"  {path} ({os.path.getsize(path)} bytes)")

def create_numbers(path):
    """Create a minimal Apple Numbers file (ZIP-based)."""
    with zipfile.ZipFile(path, 'w', zipfile.ZIP_DEFLATED) as zf:
        zf.writestr('index.xml', '<?xml version="1.0"?>\n<plist version="1.0">\n'
                    '<dict><key>DocumentType</key><string>Spreadsheet</string></dict>\n</plist>')
    print(f"  {path} ({os.path.getsize(path)} bytes)")

def create_key(path):
    """Create a minimal Apple Keynote file (ZIP-based)."""
    with zipfile.ZipFile(path, 'w', zipfile.ZIP_DEFLATED) as zf:
        zf.writestr('index.xml', '<?xml version="1.0"?>\n<plist version="1.0">\n'
                    '<dict><key>DocumentType</key><string>Presentation</string></dict>\n</plist>')
        zf.writestr('preview.jpg', b'\xFF\xD8\xFF' + b'\x00' * 50)
    print(f"  {path} ({os.path.getsize(path)} bytes)")

if __name__ == '__main__':
    print("Generating Apple format fixtures...")

    print("\n📷 Apple Images:")
    create_heic(os.path.join(FIXTURES_DIR, "sample.heic"))
    create_avif(os.path.join(FIXTURES_DIR, "sample.avif"))
    create_icns(os.path.join(FIXTURES_DIR, "sample.icns"))

    print("\n🎵 Apple Audio:")
    create_alac(os.path.join(FIXTURES_DIR, "sample.alac.m4a"))
    create_aiff(os.path.join(FIXTURES_DIR, "sample.aiff"))

    print("\n🎬 Apple Video:")
    create_prores(os.path.join(FIXTURES_DIR, "sample_prores.mov"))

    print("\n📝 Apple iWork:")
    create_pages(os.path.join(FIXTURES_DIR, "sample.pages"))
    create_numbers(os.path.join(FIXTURES_DIR, "sample.numbers"))
    create_key(os.path.join(FIXTURES_DIR, "sample.key"))

    print("\n📦 Apple Archives:")
    create_ipa(os.path.join(FIXTURES_DIR, "sample.ipa"))

    print("\n💿 Apple Disk Images:")
    create_dmg(os.path.join(FIXTURES_DIR, "sample.dmg"))

    print("\n✅ All Apple fixtures generated!")
