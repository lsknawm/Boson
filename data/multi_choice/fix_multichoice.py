import json
import os

INPUT_FILE = '../questions.json'
OUTPUT_FILE = '../questions.json'

def ensure_field(node, field, default_value):
    """原子操作：补全字段，清洗空字符串"""
    if field not in node:
        node[field] = default_value
    else:
        # 清洗空字符串为 None
        if field in ['code', 'image'] and node[field] == "":
            node[field] = None

def normalize_node(node, is_content=False):
    """通用节点清洗：补全 code, image, code_run_count 等"""
    if not isinstance(node, dict):
        return

    # 核心字段
    ensure_field(node, 'code', None)
    ensure_field(node, 'code_error', False)
    ensure_field(node, 'code_run_count', 0)  # [新增] 运行次数
    ensure_field(node, 'has_image', False)
    ensure_field(node, 'image', None)

    # 文本字段
    if 'text' not in node:
        node['text'] = ""

    # 题干调试信息
    if is_content:
        ensure_field(node, 'debug_msg', None)

def process_multiple_choice(questions):
    count = 0
    for q in questions:
        # 1. 类型过滤
        if q.get("type") != "multiple_choice":
            continue

        # 2. 修正题干 (Content)
        if "content" not in q or not isinstance(q["content"], dict):
            q["content"] = {}
        normalize_node(q["content"], is_content=True)

        # 3. 修正选项 (Structure)
        if "structure" not in q or not isinstance(q["structure"], dict):
            q["structure"] = {"layout": "vertical", "options": []}

        options = q["structure"].get("options")
        if options is None or not isinstance(options, list):
            q["structure"]["options"] = []

        for opt in q["structure"]["options"]:
            normalize_node(opt)

        # 4. 修正解析与答案 (Validation)
        if "validation" not in q or not isinstance(q["validation"], dict):
            q["validation"] = {"answer": [], "explanation": {}} # 多选题答案默认为空列表 []

        # [特有逻辑] 强制检查 answer 是否为列表，如果不是（比如是 null 或 字符串），重置为 []
        ans = q["validation"].get("answer")
        if not isinstance(ans, list):
            print(f"⚠️ 修正 ID {q.get('id')} 的 answer 类型，重置为列表 []")
            q["validation"]["answer"] = []

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

    print(f"📂 读取数据...")
    try:
        with open(INPUT_FILE, 'r', encoding='utf-8') as f:
            data = json.load(f)
    except Exception as e:
        print(f"❌ JSON 错误: {e}")
        return

    print(f"⚙️ 正在修复多选题 (multiple_choice)...")
    count = process_multiple_choice(data)

    try:
        with open(OUTPUT_FILE, 'w', encoding='utf-8') as f:
            json.dump(data, f, ensure_ascii=False, indent=2)
        print(f"✅ 完成！修正了 {count} 道多选题。")
    except Exception as e:
        print(f"❌ 保存失败: {e}")

if __name__ == "__main__":
    main()