#!/usr/bin/env python3
"""
测试数据生成脚本
为 APP UI 自动化测试生成用户数据、业务数据、异常数据
支持输出为 JSON / SQL / Dart fixture
"""

import json
import random
import string
import hashlib
import os
from datetime import datetime, timedelta
from typing import List, Dict, Any

# ── 配置 ───────────────────────────────────────────
API_BASE = "http://43.163.117.74:3000/api"
OUTPUT_DIR = os.path.join(os.path.dirname(__file__), "fixtures")

# 缅甸语名字池
MYANMAR_NAMES = [
    "Aung", "Thura", "Kyaw", "Min", "Hlaing", "Zaw", "Win", "Myat",
    "Su", "Khin", "Hla", "Than", "Soe", "Naing", "Tun", "Lwin",
    "柒少", "静牌", "超哥", "见南山", "脆皮五华", "薄注", "hch2003"
]

# 牌桌名称池
TABLE_NAMES = [
    "缅甸之星", "仰光之夜", "曼德勒风云", "内比都皇家", "蒲甘古城",
    "伊洛瓦底", "掸邦高原", "勃生港", "实皆山", "皎漂湾",
    "新手训练场", "高筹码区", "VIP 专享", "快速桌", "深筹码"
]

# 盲注级别
STAKES = [
    ("1/2", 100, 500),
    ("2/5", 250, 1250),
    ("5/10", 500, 5000),
    ("10/25", 1000, 10000),
    ("25/50", 2500, 25000),
    ("50/100", 5000, 50000),
]

# ── 工具函数 ───────────────────────────────────────

def random_id() -> str:
    return hashlib.md5(os.urandom(16)).hexdigest()[:16]

def random_date(start_days=-365, end_days=0) -> str:
    d = datetime.now() + timedelta(days=random.randint(start_days, end_days))
    return d.isoformat()

def random_phone() -> str:
    prefixes = ["09", "+95"]
    return random.choice(prefixes) + "".join(random.choices(string.digits, k=9))

def random_email(name: str) -> str:
    domains = ["gmail.com", "yahoo.com", "outlook.com", "mm.com"]
    return f"{name.lower()}{random.randint(1,999)}@{random.choice(domains)}"

# ── 1. 用户数据生成 ─────────────────────────────────

class UserGenerator:
    """生成各类用户数据"""

    @staticmethod
    def normal_user(index: int = 0) -> Dict[str, Any]:
        """正常用户"""
        name = f"testuser{index:03d}"
        return {
            "id": random_id(),
            "username": name,
            "email": random_email(name),
            "password": "Test@1234",
            "nickname": random.choice(MYANMAR_NAMES),
            "phone": random_phone(),
            "gold": random.randint(1000, 100000),
            "bb": random.randint(50, 5000),
            "status": "active",
            "created_at": random_date(-180, 0),
            "is_sim_user": False,
            "avatar": None,
        }

    @staticmethod
    def admin_user() -> Dict[str, Any]:
        """管理员用户"""
        return {
            "id": random_id(),
            "username": "admin",
            "email": "admin@depuzhiguang.com",
            "password": "Admin@8888",
            "nickname": "管理员",
            "phone": "+95123456789",
            "gold": 999999,
            "bb": 99999,
            "status": "active",
            "role": "admin",
            "created_at": random_date(-365, -300),
        }

    @staticmethod
    def edge_case_users() -> List[Dict[str, Any]]:
        """边界/异常用户"""
        return [
            {
                "id": random_id(),
                "username": "a",  # 最短用户名
                "email": "a@b.c",
                "password": "1",
                "nickname": "",
                "gold": 0,
                "bb": 0,
                "case": "minimum_values"
            },
            {
                "id": random_id(),
                "username": "x" * 50,  # 超长用户名
                "email": f"{'x'*50}@{'y'*50}.com",
                "password": "!@#$%^&*()_+~`|<>?",
                "nickname": "🔥🎰♠️♥️",
                "gold": 2147483647,  # int max
                "bb": 2147483647,
                "case": "maximum_values"
            },
            {
                "id": random_id(),
                "username": "user with spaces",  # 含空格
                "email": "not-an-email",
                "password": "   ",  # 纯空格密码
                "nickname": "<script>alert(1)</script>",  # XSS
                "gold": -100,  # 负数
                "bb": -10,
                "case": "invalid_format"
            },
            {
                "id": random_id(),
                "username": "inactive_user",
                "email": "inactive@test.com",
                "password": "Test@1234",
                "status": "suspended",
                "gold": 5000,
                "case": "suspended_user"
            },
            {
                "id": random_id(),
                "username": "duplicate_test",
                "email": "dup@test.com",
                "password": "Test@1234",
                "case": "duplicate_test"
            },
        ]

    @staticmethod
    def generate_batch(count: int = 20) -> List[Dict[str, Any]]:
        """批量生成正常用户"""
        return [UserGenerator.normal_user(i) for i in range(count)]


# ── 2. 业务数据生成 ─────────────────────────────────

class TableGenerator:
    """生成牌桌业务数据"""

    @staticmethod
    def cash_table(index: int = 0) -> Dict[str, Any]:
        stake, min_buyin, max_buyin = random.choice(STAKES)
        sb, bb = stake.split("/")
        return {
            "id": f"tbl-{random_id()}",
            "name": f"{random.choice(TABLE_NAMES)}-{index+1}",
            "type": "cash",
            "stakes": stake,
            "small_blind": int(sb),
            "big_blind": int(bb),
            "min_buyin": min_buyin,
            "max_buyin": max_buyin,
            "max_players": random.choice([6, 9, 10]),
            "current_players": random.randint(0, 10),
            "status": random.choice(["waiting", "playing", "full"]),
            "created_at": random_date(-90, 0),
        }

    @staticmethod
    def sng_table(index: int = 0) -> Dict[str, Any]:
        buyin = random.choice([10, 25, 50, 100, 500])
        return {
            "id": f"sng-{random_id()}",
            "name": f"SNG {buyin}BB 速桌-{index+1}",
            "type": "sng",
            "buyin": buyin,
            "entries": random.randint(2, 9),
            "max_entries": 9,
            "prize_pool": buyin * 9 * 0.9,
            "status": random.choice(["registering", "running", "completed"]),
            "starts_in": random.randint(0, 300),
        }

    @staticmethod
    def tournament(index: int = 0) -> Dict[str, Any]:
        return {
            "id": f"mtt-{random_id()}",
            "name": f"每日锦标赛 #{index+1}",
            "type": "tournament",
            "buyin": random.choice([50, 100, 500, 1000]),
            "guaranteed_prize": random.choice([1000, 5000, 10000, 50000]),
            "entries": random.randint(10, 200),
            "max_entries": 500,
            "start_time": random_date(0, 7),
            "status": "upcoming",
            "levels": 15,
        }

    @staticmethod
    def generate_all(cash: int = 10, sng: int = 5, mtt: int = 3) -> Dict[str, List]:
        return {
            "cash_tables": [TableGenerator.cash_table(i) for i in range(cash)],
            "sng_tables": [TableGenerator.sng_table(i) for i in range(sng)],
            "tournaments": [TableGenerator.tournament(i) for i in range(mtt)],
        }


class ClubGenerator:
    """生成俱乐部数据"""

    @staticmethod
    def club(index: int = 0) -> Dict[str, Any]:
        return {
            "id": f"club-{random_id()}",
            "name": f"俱乐部-{random.choice(MYANMAR_NAMES)}-{index+1}",
            "description": f"欢迎加入我们的俱乐部！{random.choice(['新手友好', '高额桌', '24小时开放'])}",
            "owner_id": random_id(),
            "member_count": random.randint(5, 500),
            "max_members": 1000,
            "join_type": random.choice(["approval", "open", "invite"]),
            "status": "active",
            "created_at": random_date(-365, 0),
        }

    @staticmethod
    def generate(count: int = 5) -> List[Dict[str, Any]]:
        return [ClubGenerator.club(i) for i in range(count)]


class HandHistoryGenerator:
    """生成牌局历史数据"""

    ACTIONS = ["fold", "call", "raise", "check", "all_in"]
    PHASES = ["preflop", "flop", "turn", "river"]

    @staticmethod
    def hand(table_id: str, hand_num: int = 1) -> Dict[str, Any]:
        players = random.randint(2, 10)
        return {
            "id": f"hand-{random_id()}",
            "table_id": table_id,
            "hand_number": hand_num,
            "phase": random.choice(HandHistoryGenerator.PHASES),
            "pot": round(random.uniform(0.5, 100.0), 1),
            "players_count": players,
            "actions": [
                {
                    "player": f"p{i+1}",
                    "action": random.choice(HandHistoryGenerator.ACTIONS),
                    "amount": round(random.uniform(0, 50.0), 1) if random.random() > 0.3 else 0,
                    "phase": random.choice(HandHistoryGenerator.PHASES),
                }
                for i in range(random.randint(2, 20))
            ],
            "winner": f"p{random.randint(1, players)}" if random.random() > 0.1 else None,
            "timestamp": random_date(-7, 0),
        }

    @staticmethod
    def generate_for_table(table_id: str, count: int = 50) -> List[Dict[str, Any]]:
        return [HandHistoryGenerator.hand(table_id, i+1) for i in range(count)]


# ── 3. 异常数据生成 ─────────────────────────────────

class AnomalyGenerator:
    """生成异常/边界测试数据"""

    @staticmethod
    def malformed_requests() -> List[Dict[str, Any]]:
        """畸形请求数据"""
        return [
            {"body": None, "case": "null_body"},
            {"body": {}, "case": "empty_object"},
            {"body": "not-json", "case": "invalid_json"},
            {"body": {"username": None}, "case": "null_field"},
            {"body": {"username": 12345}, "case": "wrong_type"},
            {"body": {"extra": "x" * 10000}, "case": "oversized_field"},
        ]

    @staticmethod
    def injection_attempts() -> List[Dict[str, Any]]:
        """注入攻击数据"""
        return [
            {"username": "' OR '1'='1", "case": "sql_injection"},
            {"username": "<script>alert('xss')</script>", "case": "xss"},
            {"username": "../../../etc/passwd", "case": "path_traversal"},
            {"username": "$(whoami)", "case": "command_injection"},
            {"email": "test@test.com\nBcc: attacker@evil.com", "case": "email_header_injection"},
        ]

    @staticmethod
    def race_condition_data() -> List[Dict[str, Any]]:
        """并发竞争数据"""
        base = {
            "username": "concurrent_user",
            "password": "Test@1234",
            "email": "concurrent@test.com",
        }
        return [base.copy() for _ in range(10)]

    @staticmethod
    def network_faults() -> List[Dict[str, Any]]:
        """网络故障模拟"""
        return [
            {"scenario": "timeout", "delay_ms": 30000},
            {"scenario": "partial_response", "truncate_at": 50},
            {"scenario": "connection_reset", "error": "ECONNRESET"},
            {"scenario": "dns_failure", "error": "ENOTFOUND"},
        ]


# ── 4. 输出与导出 ───────────────────────────────────

def export_json(data: Dict[str, Any], filename: str):
    """导出为 JSON fixture"""
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    path = os.path.join(OUTPUT_DIR, filename)
    with open(path, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)
    print(f"Exported: {path}")
    return path


def export_sql(users: List[Dict], tables: Dict[str, List]) -> str:
    """导出为 SQL 插入语句"""
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    path = os.path.join(OUTPUT_DIR, "test_data.sql")
    with open(path, "w", encoding="utf-8") as f:
        f.write("-- Auto-generated test data\n")
        f.write(f"-- Generated at: {datetime.now().isoformat()}\n\n")

        # Users
        for u in users:
            f.write(
                f"INSERT INTO users (id, username, email, password, nickname, gold, bb, status, created_at) "
                f"VALUES ('{u['id']}', '{u['username']}', '{u['email']}', '{u['password']}', "
                f"'{u.get('nickname', '')}', {u.get('gold', 0)}, {u.get('bb', 0)}, "
                f"'{u.get('status', 'active')}', '{u.get('created_at', datetime.now().isoformat())}');\n"
            )
        f.write("\n")

        # Tables
        for t in tables.get("cash_tables", []):
            f.write(
                f"INSERT INTO club_tables (id, club_id, name, stakes, game_type, status, created_at) "
                f"VALUES ('{t['id']}', 'club-test', '{t['name']}', '{t['stakes']}', 'cash', "
                f"'active', '{t['created_at']}');\n"
            )

    print(f"Exported SQL: {path}")
    return path


def export_dart_fixture(data: Dict[str, Any], filename: str) -> str:
    """导出为 Dart fixture 文件（用于 Flutter 测试）"""
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    path = os.path.join(OUTPUT_DIR, filename)

    dart_code = f"""// Auto-generated test fixture
// Generated at: {datetime.now().isoformat()}

const Map<String, dynamic> testUsers = {json.dumps(data.get('users', {}), ensure_ascii=False, indent=2)};

const Map<String, dynamic> testTables = {json.dumps(data.get('tables', {}), ensure_ascii=False, indent=2)};
"""
    with open(path, "w", encoding="utf-8") as f:
        f.write(dart_code)
    print(f"Exported Dart: {path}")
    return path


# ── 5. 主程序 ──────────────────────────────────────

def main():
    print("=" * 60)
    print("APP UI 测试数据生成器")
    print("=" * 60)

    # 1. 生成用户数据
    print("\n[1/5] 生成用户数据...")
    normal_users = UserGenerator.generate_batch(20)
    admin = UserGenerator.admin_user()
    edge_users = UserGenerator.edge_case_users()
    all_users = [admin] + normal_users + edge_users
    print(f"  - 正常用户: {len(normal_users)}")
    print(f"  - 管理员: 1")
    print(f"  - 边界用户: {len(edge_users)}")
    export_json({"users": all_users}, "users.json")

    # 2. 生成业务数据
    print("\n[2/5] 生成业务数据...")
    tables = TableGenerator.generate_all(cash=15, sng=8, mtt=5)
    clubs = ClubGenerator.generate(8)
    print(f"  - 现金桌: {len(tables['cash_tables'])}")
    print(f"  - SNG: {len(tables['sng_tables'])}")
    print(f"  - 锦标赛: {len(tables['tournaments'])}")
    print(f"  - 俱乐部: {len(clubs)}")
    export_json({"tables": tables, "clubs": clubs}, "tables.json")

    # 3. 生成牌局历史
    print("\n[3/5] 生成牌局历史...")
    sample_table = tables["cash_tables"][0]["id"]
    hands = HandHistoryGenerator.generate_for_table(sample_table, 30)
    print(f"  - 牌局记录: {len(hands)}")
    export_json({"hands": hands}, "hands.json")

    # 4. 生成异常数据
    print("\n[4/5] 生成异常数据...")
    anomalies = {
        "malformed": AnomalyGenerator.malformed_requests(),
        "injections": AnomalyGenerator.injection_attempts(),
        "race_condition": AnomalyGenerator.race_condition_data(),
        "network_faults": AnomalyGenerator.network_faults(),
    }
    print(f"  - 畸形请求: {len(anomalies['malformed'])}")
    print(f"  - 注入攻击: {len(anomalies['injections'])}")
    print(f"  - 并发竞争: {len(anomalies['race_condition'])}")
    print(f"  - 网络故障: {len(anomalies['network_faults'])}")
    export_json(anomalies, "anomalies.json")

    # 5. 导出 SQL
    print("\n[5/5] 导出 SQL...")
    export_sql(normal_users[:10], tables)

    # 汇总
    print("\n" + "=" * 60)
    print("数据生成完成！")
    print(f"输出目录: {OUTPUT_DIR}")
    print("=" * 60)


if __name__ == "__main__":
    main()
