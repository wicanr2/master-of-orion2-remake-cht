#!/usr/bin/env python3
# 從 openorion2 gamestate.h 抽出 C enum,解析每個成員的整數值(含隱含遞增),
# 生成 Go const 區塊(顯式值)。同時產出 markdown 對照表。
import re, sys

SRC = sys.argv[1]
OUT_GO = sys.argv[2]
OUT_MD = sys.argv[3]

text = open(SRC, encoding='utf-8', errors='replace').read()

# 已知 #define 基底值(供 enum 成員以識別字指定初值時解析)。
DEFINES = {
    'COMMON_SKILLS_TYPE': 0x0,
    'CAPTAIN_SKILLS_TYPE': 0x10,
    'ADMIN_SKILLS_TYPE': 0x20,
}

# 抓 enum Name { ... };
enum_re = re.compile(r'enum\s+([A-Za-z_]\w*)\s*\{(.*?)\}\s*;', re.DOTALL)

def strip_comments(s):
    s = re.sub(r'/\*.*?\*/', '', s, flags=re.DOTALL)
    s = re.sub(r'//[^\n]*', '', s)
    return s

go = ['// Code generated from openorion2 gamestate.h enums. DO NOT EDIT by hand.',
      '// 由 scripts/gen-enums(scratchpad genenum.py)自 openorion2 gamestate.h 生成。',
      'package gamedata', '']
md = ['# 資料枚舉對照(自 openorion2 gamestate.h 生成)', '',
      '> 由生成器自動抽出,英文名即 gameplay 邏輯 key,也是中文化術語表的基礎。', '']

for m in enum_re.finditer(text):
    name = m.group(1)
    body = strip_comments(m.group(2))
    entries = []
    counter = 0
    for part in body.split(','):
        part = part.strip()
        if not part:
            continue
        if '=' in part:
            k, v = part.split('=', 1)
            k = k.strip(); v = v.strip()
            if v in DEFINES:
                val = DEFINES[v]
            else:
                try:
                    val = int(v, 0)
                except ValueError:
                    print(f"WARN: {name}.{k} 值非整數 '{v}',跳過此 enum", file=sys.stderr)
                    entries = None
                    break
            counter = val
        else:
            k = part
            val = counter
        entries.append((k, val))
        counter += 1
    if entries is None:
        continue

    go.append(f'// {name}')
    go.append(f'type {name} int')
    go.append('const (')
    for k, val in entries:
        go.append(f'\t{k} {name} = {val}')
    go.append(')')
    go.append('')

    md.append(f'## {name}({len(entries)} 項)')
    md.append('')
    md.append('| 值 | 名稱(英) | 中文 |')
    md.append('|---|---|---|')
    for k, val in entries:
        md.append(f'| {val} | `{k}` |  |')
    md.append('')

open(OUT_GO, 'w', encoding='utf-8').write('\n'.join(go))
open(OUT_MD, 'w', encoding='utf-8').write('\n'.join(md))
print(f"生成 {OUT_GO} 與 {OUT_MD}")
