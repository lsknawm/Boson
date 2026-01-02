import json
import os

INPUT_FILE = '../questions.json'
OUTPUT_FILE = '../questions.json'

def ensure_field(node, field, default_value):
    """原子操作：补全字段，清洗空字符串"""
    if field not in node:
        node[field] = default_value
    else:
        # 清洗空字符串为 None，保持数据整洁
        if field in ['code', 'image'] and node[field] == "":
            node[field] = None

def normalize_node(node, is_content=False):
    """通用节点清洗：补全 code, image, code_run_count 等"""
    if not isinstance(node, dict):
        return

    # --- 核心字段 ---
    ensure_field(node, 'code', None)
    ensure_field(node, 'code_error', False)
    ensure_field(node, 'code_run_count', 0)  # [核心] 运行次数
    ensure_field(node, 'has_image', False)
    ensure_field(node, 'image', None)

    # --- 文本字段 ---
    if 'text' not in node:
        node['text'] = ""

    # --- 题干特有 ---
    if is_content:
        ensure_field(node, 'debug_msg', None)

def process_short_answer(questions):
    count = 0
    for q in questions:
        # 1. 类型过滤
        if q.get("type") != "short_answer":
            continue

        # 2. 修正题干 (Content)
        if "content" not in q or not isinstance(q["content"], dict):
            q["content"] = {}
        normalize_node(q["content"], is_content=True)

        # 3. 修正结构 (Structure)
        # 简答题通常没有 options，但需要保证 structure 节点存在且 layout 正确
        if "structure" not in q or not isinstance(q["structure"], dict):
            q["structure"] = {}

        # 确保 layout 存在，默认为 free_text
        if "layout" not in q["structure"]:
            q["structure"]["layout"] = "free_text"

        # 4. 修正解析与答案 (Validation)
        if "validation" not in q or not isinstance(q["validation"], dict):
            q["validation"] = {"answer": "", "explanation": {}}

        # 确保 answer 字段存在 (简答题答案通常是字符串)
        if "answer" not in q["validation"]:
            q["validation"]["answer"] = ""

        # 修正 explanation
        if "explanation" not in q["validation"] or not isinstance(q["validation"]["explanation"], dict):
            q["validation"]["explanation"] = {}
        normalize_node(q["validation"]["explanation"])

        count += 1
    return count

def main():
    if not os.path.exists(INPUT_FILE):
        print(f"❌ 找不到文件: {INPUT_FILE}")
        return

    print(f"📂 读取数据: {INPUT_FILE}...")
    try:
        with open(INPUT_FILE, 'r', encoding='utf-8') as f:
            data = json.load(f)
    except Exception as e:
        print(f"❌ JSON 格式错误: {e}")
        return

    # 支持单个对象或数组
    is_single = False
    if isinstance(data, dict):
        data = [data]
        is_single = True

    print(f"⚙️ 正在修复简答题 (short_answer)...")
    count = process_short_answer(data)

    if is_single:
        data = data[0]

    try:
        with open(OUTPUT_FILE, 'w', encoding='utf-8') as f:
            json.dump(data, f, ensure_ascii=False, indent=2)
        print(f"✅ 完成！修正了 {count} 道简答题。")
    except Exception as e:
        print(f"❌ 保存失败: {e}")

if __name__ == "__main__":
    main()