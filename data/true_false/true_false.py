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

    # --- 核心字段 ---
    ensure_field(node, 'code', None)
    ensure_field(node, 'code_error', False)
    ensure_field(node, 'code_run_count', 0)  # [核心] 运行次数计数器
    ensure_field(node, 'has_image', False)
    ensure_field(node, 'image', None)

    # --- 文本字段 ---
    if 'text' not in node:
        node['text'] = ""

    # --- 题干特有 ---
    if is_content:
        ensure_field(node, 'debug_msg', None)

def process_true_false(questions):
    count = 0
    for q in questions:
        # 1. 类型过滤 (匹配 true_false)
        if q.get("type") != "true_false":
            continue

        # 2. 修正题干 (Content)
        if "content" not in q or not isinstance(q["content"], dict):
            q["content"] = {}
        normalize_node(q["content"], is_content=True)

        # 3. 修正选项 (Structure)
        if "structure" not in q or not isinstance(q["structure"], dict):
            q["structure"] = {"layout": "horizontal", "options": []}

        options = q["structure"].get("options")

        # 即使是 T/F，也需要补全字段
        if isinstance(options, list):
            for opt in options:
                normalize_node(opt)
        else:
            # 如果 options 为空或非列表，初始化为空列表 (或者你可以在这里生成默认的 T/F 选项)
            q["structure"]["options"] = []

        # 4. 修正解析与答案 (Validation)
        if "validation" not in q or not isinstance(q["validation"], dict):
            q["validation"] = {"answer": "", "explanation": {}}

        # 确保 explanation 节点完整
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

    # 如果 data 是单个对象而不是列表，临时转为列表处理
    is_single = False
    if isinstance(data, dict):
        data = [data]
        is_single = True

    print(f"⚙️ 正在修复 True/False 题目...")
    count = process_true_false(data)

    # 如果输入是单个对象，还原回去
    if is_single:
        data = data[0]

    try:
        with open(OUTPUT_FILE, 'w', encoding='utf-8') as f:
            json.dump(data, f, ensure_ascii=False, indent=2)
        print(f"✅ 完成！修正了 {count} 道题目。")
    except Exception as e:
        print(f"❌ 保存失败: {e}")

if __name__ == "__main__":
    main()