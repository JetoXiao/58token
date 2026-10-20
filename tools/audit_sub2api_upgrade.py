"""Read-only Git snapshot audit; writes review artifacts, never merges or migrates."""
import hashlib
import json
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / 'docs' / 'upgrade-v0.2.17'
BASE = '63b0631a58272e88af0d83aa168f65787c53a746'
INITIAL = '75b6c3c5910b4bb525f0337bbd36ae29911f7dc8'
LOCAL = 'f2f785fd6af5233f570841003fb3fb23d830d78b'
TARGET = '060d0160417782cd962a0c7f460f6cd4565d4916'


def git(*args):
    return subprocess.check_output(['git', *args], cwd=ROOT)


def tree(ref):
    result = {}
    for row in git('ls-tree', '-rz', ref).split(b'\0'):
        if row:
            info, path = row.split(b'\t', 1)
            result[path.decode('utf-8')] = info.split()[2].decode()
    return result


def domain(path):
    if '/migrations/' in path or '/ent/' in path:
        return 'database'
    for key, words in [
        ('payment-subscription-affiliate', ['payment', 'subscription', 'affiliate', 'redeem', 'trial']),
        ('marketplace-pricing', ['marketplace', 'pricing', 'upstream_balance']),
        ('response-cache', ['response_cache']),
        ('gateway-scheduler', ['gateway', 'failover', 'runtime_', 'scheduler', 'images', 'ratelimit', 'sticky']),
        ('auth-permissions', ['auth', 'permission', 'totp', 'jwt']),
        ('content-analytics', ['visitor', 'help_center', 'download', 'support_contact']),
    ]:
        if any(word in path.lower() for word in words):
            return key
    if path.startswith('frontend/'):
        return 'frontend-shared'
    if path.startswith(('deploy/', '.github/')) or 'Dockerfile' in path or path.endswith(('go.mod', 'go.sum', 'Makefile')):
        return 'build-deploy'
    return 'shared-other'


def main():
    OUT.mkdir(parents=True, exist_ok=True)
    trees = {key: tree(ref) for key, ref in [('base', BASE), ('initial', INITIAL), ('local', LOCAL), ('target', TARGET)]}
    initial_diff = [p for p in sorted(set(trees['base']) | set(trees['initial'])) if trees['base'].get(p) != trees['initial'].get(p)]
    files = []
    for p in sorted(set(trees['base']) | set(trees['local']) | set(trees['target'])):
        b, l, t = (trees[k].get(p) for k in ['base', 'local', 'target'])
        if b == l == t:
            continue
        status = 'same-result' if l == t else 'upstream-only-change' if l == b else 'custom-only-change' if t == b else 'both-changed-review'
        files.append(dict(path=p, domain=domain(p), classification=status, base_blob=b, local_blob=l, target_blob=t))
    migrations = []
    paths = sorted(p for p in set(trees['local']) | set(trees['target']) if p.startswith('backend/migrations/') and p.endswith('.sql'))
    for p in paths:
        contents = {k: git('cat-file', 'blob', trees[k][p]).decode('utf-8') if p in trees[k] else None for k in ['local', 'target']}
        l, t = contents['local'], contents['target']
        status = 'local-only-preserve' if t is None else 'target-only-review' if l is None else 'same' if l == t else 'different'
        # A lexical hint only; dollar-quoted SQL/string literals require human review.
        strip_comments = lambda s: re.sub(r'/\*.*?\*/|--[^\n]*', '', s, flags=re.S).strip()
        comment_hint = status == 'different' and strip_comments(l) == strip_comments(t)
        hashes = {k: hashlib.sha256(v.strip().encode()).hexdigest() if v is not None else None for k, v in contents.items()}
        statements = [line.strip()[:220] for line in (t or l).splitlines() if re.match(r'^\s*(CREATE|ALTER|DROP|UPDATE|DELETE|INSERT|TRUNCATE)\b', line, re.I)]
        migrations.append(dict(filename=Path(p).name, classification=status, comment_only_hint=comment_hint,
                               runner_sha256=hashes, statement_hints=statements,
                               production_checksum_verified=False,
                               action='Keep published local bytes; inspect target initialization semantics' if comment_hint else 'Preserve local history' if t is None else 'Review dependencies, destructive changes and existing schema' if l is None else 'Keep identical' if status == 'same' else 'Manual reconciliation required'))
    counts = {s: sum(f['classification'] == s for f in files) for s in sorted({f['classification'] for f in files})}
    summary = dict(base=BASE, initial=INITIAL, local=LOCAL, target=TARGET, initial_differences=initial_diff,
                   tracked_files={k: len(v) for k, v in trees.items()}, file_classifications=counts,
                   migration_classifications={s: sum(m['classification'] == s for m in migrations) for s in sorted({m['classification'] for m in migrations})})
    for name, value in [('summary.json', summary), ('file-inventory.json', files), ('migration-inventory.json', migrations)]:
        (OUT / name).write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    lines = ['# 迁移逐文件审查清单', '', '自动生成：`python tools/audit_sub2api_upgrade.py`。哈希按 UTF-8 内容 trim 后计算，与当前 runner 规则对照。生产数据库记录尚未获取，源码哈希不代表生产已应用状态。', '',
             '“注释差异”是词法提示；不是 SQL 语义等价证明。下表未核实生产 schema、锁时长、执行结果。', '',
             '| 文件 | 分类 | 建议处理 |', '| --- | --- | --- |']
    for m in migrations:
        if m['classification'] == 'same':
            continue
        action = '保留本项目原字节；目标仅注释差异提示' if m['comment_only_hint'] else '保留已发布定制，审查上游结构交互' if m['classification'].startswith('local-only') else '审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证'
        lines.append(f"| `{m['filename']}` | {m['classification']} | {action} |")
    (OUT / 'MIGRATION_REVIEW.md').write_text('\n'.join(lines) + '\n', encoding='utf-8')
    print(json.dumps(summary, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    main()
